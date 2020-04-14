package system
import(
	"github.com/carlosfrodrigues/bittorrent/bencoder"
	"io/ioutil"
	"strings"
	"crypto/sha1"
	"net/url"
	"math/rand"
	"strconv"
	"os"
)

type MetaData struct{
	Name string
	InfoHash []byte
	Length int
	PieceLength int
	Pieces [][]byte
	Announce string
	Url string
	PeerId string

}

func (d *MetaData) Init(fileName string) {
	dat, _ := ioutil.ReadFile(fileName)
	mapa, _ := bencoder.Decode(dat)
	torrentDict := mapa.(map[string]interface{})
	//fmt.Println(reflect.TypeOf())
	d.Name = string(torrentDict["info"].(map[string]interface{})["name"].([]byte))
	d.InfoHash = generateHash(dat)
	d.Length = int(torrentDict["info"].(map[string]interface{})["length"].(int64))
	d.PieceLength = int(torrentDict["info"].(map[string]interface{})["piece length"].(int64))
	d.PeerId = "-PC0001-" + strconv.Itoa(100000000000 + rand.Intn(899999999999))
	pieces := torrentDict["info"].(map[string]interface{})["pieces"].([]byte)	
	piecesLen := len(pieces)
	offset := 0
	for offset < piecesLen {
		d.Pieces = append(d.Pieces, pieces[offset:offset+20])
		offset = offset+20
	}
	
	d.Announce = string(torrentDict["announce"].([]byte))

	GenerateUrl(d)
}


func GenerateUrl(d *MetaData) {
	params := url.Values{}
	params.Add("info_hash", string(d.InfoHash))
	params.Add("peer_id", d.PeerId)
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.Itoa(d.Length))
	params.Add("port", "6889")
	params.Add("compact", "1")
	url := d.Announce + "?" + params.Encode()
	d.Url = url
}

func generateHash(bencodeData []byte) []byte{
	dataToHash := strings.Split(string(bencodeData[:]),"4:info")[1]
	h := sha1.New()
	h.Write([]byte(dataToHash[:len(dataToHash)-1]))
	bs := h.Sum(nil)
	return bs
}

func (d *MetaData) DownloadFile(path string, peers []Peer) error {
	torrent := Torrent{
		Peers: peers,
		PeerID: d.PeerId,
		InfoHash: d.InfoHash,
		Pieces: d.Pieces,
		PieceLength: d.PieceLength,
		Length: d.Length,
		Name: d.Name,
		Url: d.Url,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}