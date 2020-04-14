package main
import (
	"os"
	"github.com/carlosfrodrigues/bittorrent/system"
	"fmt"
)

func main(){
	if(len(os.Args) == 1){
		fmt.Println("Usage: program <torrentfile>")
		return
	}
	inFile := os.Args[1]
	meta := new(system.MetaData)
	meta.Init(inFile)
	request := <-system.GetResponseTracker(meta.Url)
	meta.DownloadFile(meta.Name, request.Peers)
	
}
