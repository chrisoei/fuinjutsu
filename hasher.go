package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
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

func storedData(data []byte) []byte {
	if len(data) < 256 {
		return data
	}
	return nil
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

func hashFile(filename string) (map[string]string, []byte) {
	h := multidigest.New()
	w := h.Writer()
	f, err := os.Open(filename)
	defer f.Close()
	oei.ErrorHandler(err)

	// grab the first 512 bytes
	data := make([]byte, 512)
	f.Read(data)
	// send it to the hash function
	w.Write(data)

	// then copy any remaining bytes in bulk
	io.Copy(w, f)

	// we grabbed more than 256 bytes above so storedData knows if it needs to truncate it here
	return h.Result(), storedData(data)
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

		r, data := hashFile(flag.Arg(n))

		if oei.Verbosity() >= 0 {
			s, err := json.MarshalIndent(r, "", "  ")
			oei.ErrorHandler(err)
			fmt.Printf("%s\n", string(s))
		}

		_, err := db.Exec(`INSERT INTO hashes(bytes, adler32, crc32, md5, ripemd160, sha1, "sha2-256", "sha2-512", "sha3-256", ssdeep29, size, version) SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12 WHERE NOT EXISTS (SELECT "sha2-256", "sha3-256" FROM hashes where "sha2-256" = $7 AND "sha3-256" = $9)`,
			storedData(data), // 1
			r["adler32"],     // 2
			r["crc32"],       // 3
			r["md5"],         // 4
			r["ripemd160"],   // 5
			r["sha1"],        // 6
			r["sha2-256"],    // 7
			r["sha2-512"],    // 8
			r["sha3-256"],    // 9
			r["ssdeep29"],    // 10
			r["size"],        // 11
			r["version"])     // 12
		oei.ErrorHandler(err)

		row := db.QueryRow(`SELECT id FROM hashes WHERE "sha2-256" = $1 AND "sha3-256" = $2`, r["sha2-256"], r["sha3-256"])
		var z int64
		row.Scan(&z)
		if oei.Verbosity() >= 0 {
			fmt.Printf("%d\n", z)
		}

		addAnnotation(db, z, "filename", &fn)
		addAnnotation(db, z, "comment", comment)
		addAnnotation(db, z, "url", url)
		addProperty(db, z, "rating", rating)
		addProperty(db, z, "imdb", imdb)
		addTag(db, z, tag)
	}
}
