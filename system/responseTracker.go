package system
import(
	"encoding/binary"
	"net"
	"net/http"
	"fmt"
	"io/ioutil"
	"github.com/carlosfrodrigues/bittorrent/bencoder"
	"strconv"

)
type Peer struct{
	IP net.IP
	Port uint16
}

type responseTracker struct{
	Complete int
	Incomplete int
	Interval int
	Peers []Peer
}

func GetResponseTracker(url string) <-chan *responseTracker{
	r := make(chan *responseTracker)
	go func(){
		defer close(r)
		res, err := http.Get(url)
		if(err != nil){
			fmt.Println("error on getting response tracker")
			return
		}

		defer res.Body.Close()
		content, err := ioutil.ReadAll(res.Body)
		resTrack := new(responseTracker)
		mapa, _ := bencoder.Decode([]byte(content))
		resTrack.Interval = int(mapa.(map[string]interface{})["interval"].(int64))
		peersCompressed := mapa.(map[string]interface{})["peers"].([]byte)
		sizePeers := len(peersCompressed)
		numPeers := sizePeers/6
		peers := make([]Peer, numPeers)
		for i := 0; i < sizePeers; i+=6{
			peers[i/6].IP = net.IP(peersCompressed[i:i+4])
			peers[i/6].Port = binary.BigEndian.Uint16(peersCompressed[i+4:i+6])
		}
		resTrack.Peers = peers
		r <- resTrack
	}()
	
	return r
}

func (p Peer) ToString() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}