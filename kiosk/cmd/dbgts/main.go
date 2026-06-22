package main
import ("database/sql";"fmt";_ "modernc.org/sqlite")
func main(){
 db,_:=sql.Open("sqlite","file:data/kiosk.db?mode=ro");defer db.Close()
 var ph string
 db.QueryRow(`SELECT coalesce(placeholders,'(null)') FROM surat_template WHERE id='7a553a02-9d30-4fa2-ba4f-68d8bccd9424'`).Scan(&ph)
 fmt.Println(ph)
}
