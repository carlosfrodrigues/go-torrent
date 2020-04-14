package main
import (
	"fmt"
	"os"
	"github.com/carlosfrodrigues/bittorrent/system"
)

func main(){
	inFile := os.Args[1]
	meta := new(system.MetaData)
	meta.Init(inFile)
	request := <-system.GetResponseTracker(meta.Url)
	meta.DownloadFile(meta.Name, request.Peers)
	
}
