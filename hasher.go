package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chrisoei/multidigest"
	"github.com/chrisoei/oei"
	"github.com/chrisoei/xattr"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func getDb() *sql.DB {
	url := os.Getenv("HASHDB")
	if oei.Verbosity() > 10 {
		log.Println("URL: " + url)
	}
	cs, err := pq.ParseURL(url)
	oei.ErrorHandler(err)
	if oei.Verbosity() > 10 {
		log.Println("Connection string: " + cs)
	}
	db, err := sql.Open("postgres", cs)
	oei.ErrorHandler(err)
	return db
}

func addAnnotation(db *sql.DB, hash_id int64, t string, a *string) {
	if *a != "" {
		_, err := db.Exec(`INSERT INTO annotations(hash_id, type, annotation) VALUES($1, $2, $3)`, hash_id, t, a)
		oei.ErrorHandler(err)
	}
}

func addProperty(db *sql.DB, hash_id int64, t string, p *string) {
	if *p != "" {
		_, err := db.Exec(`INSERT INTO properties(hash_id, type, property) VALUES($1, $2, $3)`, hash_id, t, p)
		oei.ErrorHandler(err)
	}
}

func addTag(db *sql.DB, hash_id int64, t *string) {
	if *t != "" {
		_, err := db.Exec(`INSERT INTO tags(hash_id, tag) VALUES($1, $2)`, hash_id, t)
		oei.ErrorHandler(err)
	}
}

func addContents(db *sql.DB, hash_id int64, t []byte) {
	if len(t) > 0 {
		_, err := db.Exec(`INSERT INTO contents(hash_id, bytes) VALUES($1, $2)`, hash_id, t)
		oei.ErrorHandler(err)
	}
}

func getHashId(filename string) string {
	rg := regexp.MustCompile("\\[#(\\d+)\\]\\.\\w+")
	hidF := ""
	if rg.Match([]byte(filename)) {
		hidF = rg.FindStringSubmatch(filename)[1]
	}
	hidXbytes, err := xattr.Get(filename, "user.io.oei.hash_id")
	hidX := string(hidXbytes)
	if err == nil {
		if hidF != "" && hidF != hidX {
			return ""
		} else {
			return hidX
		}
	} else {
		return hidF
	}
}

func hashFile(filename string, save bool) (map[string]string, []byte) {
	h := multidigest.New()
	w := h.Writer()
	f, err := os.Open(filename)
	defer f.Close()
	oei.ErrorHandler(err)

	if save {
		data, err := ioutil.ReadFile(filename)
		oei.ErrorHandler(err)
		w.Write(data)
		return h.Result(), data
	}

	io.Copy(w, f)
	return h.Result(), nil
}

func hashName(filename string, hash_id int64) string {
	r := regexp.MustCompile("\\[#\\d+\\]\\.\\w+")
	if r.Match([]byte(filename)) {
		return filename
	}
	return fmt.Sprintf("%s_[#%d]%s", oei.FilenameWithoutExt(filename), hash_id, filepath.Ext(filename))
}

func main() {
	db := getDb()
	defer db.Close()

	var fn string
	var fnOverride = flag.String("file", "", "Filename to save (* for auto)")
	var comment = flag.String("comment", "", "Comment")
	var url = flag.String("url", "", "URL")
	var rating = flag.String("rating", "", "Rating")
	var imdb = flag.String("imdb", "", "IMDB")
	var tag = flag.String("tag", "", "Tag")

	var save = flag.Bool("save", false, "Save file in database")
	var useXattr = flag.Bool("xattr", false, "Save hash ID in extended attributes")
	var rename = flag.Bool("rename", false, "Rename the file using the hash ID as part of the filename")
	var verify = flag.Bool("verify", false, "Verify that the contents of the file have not changed")

	flag.Parse()

	for n := range flag.Args() {

		filename := flag.Arg(n)

		r, data := hashFile(filename, *save)

		if oei.Verbosity() >= 0 {
			s, err := json.MarshalIndent(r, "", "  ")
			oei.ErrorHandler(err)
			fmt.Printf("%s\n", string(s))
		}

		if *verify {
			hid := getHashId(filename)
			if hid == "" {
				fmt.Printf("UNKNOWN: %s\n", filename)
			} else {
				row := db.QueryRow(`SELECT "sha2-256" FROM hashes where id = $1`, hid)
				var sha256 string
				row.Scan(&sha256)
				if sha256 == r["sha2-256"] {
					fmt.Printf("OK: %s\n", filename)
				} else {
					fmt.Printf("ERROR: %s\n", filename)
				}
			}
		} else {
			_, err := db.Exec(`INSERT INTO hashes(adler32, crc32, md5, ripemd160, sha1, "sha2-256", "sha2-512", "sha3-256", ssdeep29, size, version) SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 WHERE NOT EXISTS (SELECT "sha2-256", "sha3-256" FROM hashes where "sha2-256" = $6 AND "sha3-256" = $8)`,
				r["adler32"],   // 1
				r["crc32"],     // 2
				r["md5"],       // 3
				r["ripemd160"], // 4
				r["sha1"],      // 5
				r["sha2-256"],  // 6
				r["sha2-512"],  // 7
				r["sha3-256"],  // 8
				r["ssdeep29"],  // 9
				r["size"],      // 10
				r["version"])   // 11
			oei.ErrorHandler(err)

			row := db.QueryRow(`SELECT id FROM hashes WHERE "sha2-256" = $1 AND "sha3-256" = $2`, r["sha2-256"], r["sha3-256"])
			var z int64
			row.Scan(&z)
			hid := fmt.Sprintf("%d", z)
			if *useXattr {
				oei.ErrorHandler(xattr.Set(filename, "user.io.oei.hash_id", []byte(hid)))
			}
			if oei.Verbosity() >= 0 {
				fmt.Println(hid)
			}
			if *rename {
				newFilename := hashName(filename, z)
				if newFilename != filename {
					os.Rename(filename, newFilename)
					filename = newFilename
				}
			}

			switch *fnOverride {
			case "*":
				{
					fn = filename
				}
			case "":
				{
					fn = ""
				}
			default:
				{
					fn = *fnOverride
				}
			}

			addAnnotation(db, z, "filename", &fn)
			addAnnotation(db, z, "comment", comment)
			addAnnotation(db, z, "url", url)
			addProperty(db, z, "rating", rating)
			addProperty(db, z, "imdb", imdb)
			addTag(db, z, tag)
			addContents(db, z, data)
		}
	}
}
