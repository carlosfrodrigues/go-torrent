package main
import (
	"fmt"
	"os"
	"github.com/carlosfrodrigues/bittorrent/system"
)

func main(){
	inFile := os.Args[1]
	fmt.Println(inFile)
	outFile := os.Args[2]
	//mapa := make(map[string]interface{}{})
	//fmt.Println(reflect.TypeOf(inFile))
	meta := new(system.MetaData)
	meta.Init("debian-10.3.0-amd64-DVD-1.iso.torrent")
	request := <-system.GetResponseTracker(meta.Url)
	meta.DownloadFile(outFile, request.Peers)
	//fmt.Println(r.Peers)
	//fmt.Println(system.MakeHandshake(meta.GenerateHandshake()))
	
}
