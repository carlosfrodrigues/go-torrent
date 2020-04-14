package system
import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Torrent struct {
	Peers []Peer
	PeerID string
	InfoHash []byte
	Pieces [][]byte
	PieceLength int
	Length int
	Name string
	Url string
}


type pieceWork struct {
	index int
	hash []byte
	length int
}

type pieceResult struct {
	index int
	buf []byte
}

type pieceProgress struct {
	index int
	client *Client
	buf []byte
	downloaded int
	requested int
	backlog int
}

//Largest number of bytes a request can ask for
const MaxBlockSize = 16384

//Number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

func (t *Torrent) Download() ([]byte, error) {
	//log.Println("Starting download for", t.Name)
	workQueue := make(chan *pieceWork, len(t.Pieces))
	results := make(chan *pieceResult)

	//interface
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := widgets.NewList()
	l.TextStyle = ui.NewStyle(ui.ColorRed)
	l.WrapText = false
	l.Title = "Log"
	l.SetRect(0, 0, 100, 8)
	l.Rows = []string{
		"Starting...",
	}
	g := widgets.NewGauge()
	g.Title = "Download"
	g.SetRect(0, 11, 50, 14)
	g.Percent = 0
	g.BarColor = ui.ColorBlue
	g.Label = fmt.Sprintf("%v%% (Downloading)", g.Percent)

	ui.Render(g, l)
	//interface end

	for index, hash := range t.Pieces {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}
	/*
	interval := 30000000 * time.Microsecond
	timeRefresh := time.Now().Add(interval)
	*/
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.Pieces) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.Pieces)) * 100
		numWorkers := runtime.NumGoroutine() - 1
		//log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
		//interface		
		g.Percent = int(percent)
		g.Label = fmt.Sprintf("%v%% (Downloading)", percent)
		result := fmt.Sprintf("(%0.2f%%) Downloaded piece #%d from %d peers", percent, res.index, numWorkers)
		l.Rows = append(l.Rows, result)
		if(len(l.Rows) > 6){
			l.Rows = l.Rows[1:]
		}
		ui.Render(g, l)
		//interface end
		/*
		if !(time.Now().Before(timeRefresh)) {
			newPeers := refreshPeers(t.Url, t.Peers)
			//insert new workers with the new Peers
			for _, peer := range newPeers {
				go t.startDownloadWorker(peer, workQueue, results)
			}
            timeRefresh = time.Now().Add(interval)
        }*/
	}
	close(workQueue)

	return buf, nil
}

func refreshPeers(url string, currentPeers []Peer) []Peer {
	request := <-GetResponseTracker(url)
	var newPeers []Peer
	for _, peer := range request.Peers{
		found := false 
		for _, peerCurrent := range currentPeers{
			if(peer.ToString() == peerCurrent.ToString()){
				found = true
			}
		}
		if(found == false){
			newPeers = append(newPeers, peer)
		}
	}

	return newPeers
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) startDownloadWorker(peer Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	c, err := NewConnection(peer, t.PeerID, t.InfoHash)
	if err != nil {
		//log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	//log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw 
			continue
		}


		buf, err := downloadPiece(c, pw)
		if err != nil {
			//log.Println("Exiting", err)
			workQueue <- pw 
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			//log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func downloadPiece(c *Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case MsgUnchoke:
		state.client.Choked = false
	case MsgChoke:
		state.client.Choked = true
	case MsgHave:
		index, err := ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case MsgPiece:
		n, err := ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}