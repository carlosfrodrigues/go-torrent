package system
import(
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"fmt"
	"log"
)

func startUi(){
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := widgets.NewList()
	l.TextStyle = ui.NewStyle(ui.ColorRed)
	l.WrapText = false
	l.SetRect(0, 0, 100, 8)
	l.Rows = []string{
		"Starting...",
	}
	g := widgets.NewGauge()
	g.Title = "Download"
	g.SetRect(0, 11, 50, 14)
	g.Percent = 0
	g.BarColor = ui.ColorBlue
	g.Label = fmt.Sprintf("%v%% (100MBs free)", g.Percent)

	ui.Render(g, l)
}
/*
func updateUi(percent float64, index, numWorkers int){
	g.Percent = percent
	g.Label = fmt.Sprintf("%v%% (100MBs free)", g.Percent)
	result := fmt.Sprintf("(%0.2f%%) Downloaded piece #%d from %d peers", percent, index, numWorkers)
	l.Rows = append(l.Rows, result)
	if(len(l.Rows) > 6){
		l.Rows = l.Rows[1:]
	}
	ui.Render(g, l)
}*/