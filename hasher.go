package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/AlekSi/xattr"
	"github.com/chrisoei/multidigest"
	"github.com/chrisoei/oei"
	_ "github.com/lib/pq"
	"github.com/lib/pq"
	"io"
	"io/ioutil"
	"log"
	"os"
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

	flag.Parse()

	for n := range flag.Args() {

		switch *fnOverride {
		case "*":
			{
				fn = flag.Arg(n)
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

		r, data := hashFile(flag.Arg(n), *save)

		if oei.Verbosity() >= 0 {
			s, err := json.MarshalIndent(r, "", "  ")
			oei.ErrorHandler(err)
			fmt.Printf("%s\n", string(s))
		}

		_, err := db.Exec(`INSERT INTO hashes(bytes, adler32, crc32, md5, ripemd160, sha1, "sha2-256", "sha2-512", "sha3-256", ssdeep29, size, version) SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12 WHERE NOT EXISTS (SELECT "sha2-256", "sha3-256" FROM hashes where "sha2-256" = $7 AND "sha3-256" = $9)`,
			data,           // 1
			r["adler32"],   // 2
			r["crc32"],     // 3
			r["md5"],       // 4
			r["ripemd160"], // 5
			r["sha1"],      // 6
			r["sha2-256"],  // 7
			r["sha2-512"],  // 8
			r["sha3-256"],  // 9
			r["ssdeep29"],  // 10
			r["size"],      // 11
			r["version"])   // 12
		oei.ErrorHandler(err)

		row := db.QueryRow(`SELECT id FROM hashes WHERE "sha2-256" = $1 AND "sha3-256" = $2`, r["sha2-256"], r["sha3-256"])
		var z int64
		row.Scan(&z)
		hid := fmt.Sprintf("%d", z)
		if *useXattr {
			oei.ErrorHandler(xattr.Set(flag.Arg(n), "user.io.oei.hash_id", []byte(hid)))
		}
		if oei.Verbosity() >= 0 {
			fmt.Println(hid)
		}

		addAnnotation(db, z, "filename", &fn)
		addAnnotation(db, z, "comment", comment)
		addAnnotation(db, z, "url", url)
		addProperty(db, z, "rating", rating)
		addProperty(db, z, "imdb", imdb)
		addTag(db, z, tag)
	}
}
