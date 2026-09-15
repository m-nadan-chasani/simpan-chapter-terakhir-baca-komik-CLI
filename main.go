package main

import (
	"fmt"
	"strings"
	"os"
	"os/exec"
	"runtime"
	"bufio"
	"unicode/utf8"
	"database/sql"
	_ "modernc.org/sqlite"
)

const (
    GARIS = "========================================"
    TITLE = "            COMIC HISTORIES"
    VERSION = "v2.2"
)

type Comic struct {
	ID      int
	Name    string
	Chapter string
}

var (
    reader *bufio.Reader = bufio.NewReader(os.Stdin)
	cmd *exec.Cmd
)

func clearScr() {
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func firstChar(str string) rune {
    r, _ := utf8.DecodeRuneInString(str)
    return r
}

func readAll(comic *Comic, db *sql.DB){
    var (
        err error
        rows *sql.Rows
    )
    
    clearScr()
    
	rows, err = db.Query(`
		SELECT * FROM comic_histories;
	`)
    if err != nil {
	    panic(err)
	}
	defer rows.Close()
	
	fmt.Println(GARIS)
	for rows.Next() {
        err = rows.Scan(&comic.ID, &comic.Name, &comic.Chapter)
        if err != nil {
            panic(err)
        }
        
        fmt.Println("ID     :", comic.ID)
        fmt.Println("Name   :", comic.Name)
        fmt.Println("Chapter:", comic.Chapter)
        fmt.Println(GARIS)
	}
} // end of func readAll

func addOrUpdateChapter(comic *Comic, db *sql.DB){
    var (
        err error
        row *sql.Row
        oldChapter string
    )
    
    fmt.Print("Nama    :")
    comic.Name, _ = reader.ReadString('\n')
    comic.Name = strings.TrimSpace(comic.Name)
    if comic.Name == "" {return}
    comic.Name = strings.ToLower(comic.Name)
	
    fmt.Print("Chapter :")
    comic.Chapter, _ = reader.ReadString('\n')
    comic.Chapter = strings.TrimSpace(comic.Chapter)
    if comic.Chapter == "" {return}
	
	row = db.QueryRow(`
		SELECT ID, chapter FROM comic_histories
		WHERE Name=? LIMIT 1;
	`, comic.Name)
	if err = row.Scan(&comic.ID, &oldChapter); err != nil {
        if err == sql.ErrNoRows {
            _, err = db.Exec(`
                INSERT INTO comic_histories (Name, Chapter)
                VALUES (?, ?);
            `, comic.Name, comic.Chapter)
            clearScr()
            fmt.Println("<=", comic.Name, "telah ditambahkan [chapter", comic.Chapter+"] =>")
            return
        }
        panic(err)
	}
	
	_, err = db.Exec(`
        UPDATE comic_histories
        SET Chapter=?
        WHERE ID=?;
	`, comic.Chapter, comic.ID)
	if err != nil {
        panic(err)
	} else {
        clearScr()
        fmt.Println("<= sukses update chapter ["+ oldChapter, "->", comic.Chapter+"] =>")
	}
} // akhir dari fungsi insertOrReplace

func searchComic(comic *Comic, db *sql.DB){
    var (
        err error
        rows *sql.Rows
    )
    
    fmt.Print("Nama:")
    comic.Name, _ = reader.ReadString('\n')
    comic.Name = strings.TrimSpace(comic.Name)
    if comic.Name == "" {return}
    comic.Name = strings.ToLower(comic.Name)
    
    clearScr()
    
	rows, err = db.Query(`
		SELECT * FROM comic_histories WHERE name LIKE ?;
	`, "%"+comic.Name+"%")
    if err != nil {
	    panic(err)
	}
	defer rows.Close()
	
	found := false
	fmt.Println(GARIS)
	for rows.Next() {
        err = rows.Scan(&comic.ID, &comic.Name, &comic.Chapter)
        if err != nil {
            panic(err)
        }
        
        found = true
        fmt.Println("ID     :", comic.ID)
        fmt.Println("Name   :", comic.Name)
        fmt.Println("Chapter:", comic.Chapter)
        fmt.Println(GARIS)
	}
	if !found {
	    fmt.Println("<= tidak ditemukan satupun =>")
	}
} // akhir dari fungsi searchComic

func renameComicById(comic *Comic, db *sql.DB){
    var (
        err error
        row *sql.Row
    )
    
    for {
        fmt.Print("ID   :")
        _, err = fmt.Scan(&comic.ID)
        if err != nil {
            continue
        }
        break
    }
    if comic.ID < 0 {return}
    
    row = db.QueryRow(`
		SELECT name FROM comic_histories
		WHERE ID=?
		LIMIT 1;
	`, comic.ID)
	if err = row.Scan(&comic.Name); err != nil {
        if err == sql.ErrNoRows {
            clearScr()
            fmt.Println("<= id:", comic.ID, "tidak ditemukan =>")
            return
        }
		panic(err)
	}
	
	fmt.Printf("Name :%s\n", comic.Name)
	fmt.Print("New Name :")
	comic.Name, _ = reader.ReadString('\n')
    comic.Name = strings.TrimSpace(comic.Name)
    if comic.Name == "" {return}
    comic.Name = strings.ToLower(comic.Name)
	
	_, err = db.Exec(`
		UPDATE comic_histories
		SET name=?
		WHERE ID=?
	;`, comic.Name, comic.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("<= nama telah diubah =>")
}// akhir dari fungsi renameComicById

func main(){
    clearScr()
    
	db, err := sql.Open("sqlite","comic_histories.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	
	if err = db.Ping(); err != nil {
        fmt.Println("database tidak terkoneksi")
        panic(err)
    }

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comic_histories (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			chapter TEXT NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("database berhasil dibuka")
	}

	var (
		pilihan string
		comic Comic
	)
	fmt.Println(GARIS)
	fmt.Println(TITLE)
	for exit:=false;!exit; {
		fmt.Println(GARIS)
		fmt.Println("1. tampilkan semua komik")
		fmt.Println("2. tambah/perbarui chapter komik")
		fmt.Println("3. cari komik")
		fmt.Println("4. namai ulang komik dengan id")
		fmt.Println("-. bersihkan layar")
		fmt.Println("v. version")
		fmt.Println("0. exit")
		fmt.Println(GARIS)
		fmt.Print("pilih ->")
		fmt.Scanln(&pilihan)
		if pilihan == "" {
            fmt.Println("pilihan tidak boleh kosong")
            continue
		}
		
		switch firstChar(pilihan) {
			case '1':
				readAll(&comic, db)
				
			case '2':
			    fmt.Println("kosongkan input untuk membatalkan")
				addOrUpdateChapter(&comic, db)
				
			case '3':
			    fmt.Println("kosongkan input untuk membatalkan")
                searchComic(&comic, db)
                
            case '4':
                fmt.Println("masukkan angka negatif untuk membatalkan")
                renameComicById(&comic, db)
                
            case '-':
                clearScr()
                fmt.Println(GARIS)
                fmt.Println(TITLE)
            
            case 'v':
                clearScr()
                fmt.Println("version =", VERSION)
                
			case '0':
				exit = true
				
			default:
				fmt.Println("tidak ada dalam pilihan")
		}
	}

	fmt.Println("Program selesai")
}
