package worker

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"net/http"
)

var covid19Handler CommandHandler = func(w *Worker, command string, subCommand string, args []string, evt *model.MessageEvent) error {
	resp, err := http.DefaultClient.Get("https://ncov.moh.gov.vn/")
	if err != nil {
		log.Println("Fail to get covid19 error")
		return err
	}
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println()
		return err
	}

	lines := ""

	document.Find("div.box-tke > div").Each(func(_ int, div *goquery.Selection) {
		line := ""
		div.Find("span").Each(func(spanIndex int, s *goquery.Selection) {
			switch spanIndex {
			case 0:
				line = line + "<b>" + s.Text() + "</b>\n"
				break
			case 1:
				line = line + "  - <b>Số ca nhiễm:</b>\t\t\t" + s.Text() + "\n"
				break
			case 2:
				line = line + "  - Đang Điều Trị:\t\t\t" + s.Text() + "\n"
				break
			case 3:
				line = line + "  - Đã Khỏi:\t\t\t\t" + s.Text() + "\n"
				break
			case 4:
				line = line + "  - <b>Tử Vong:</b>\t\t\t\t" + s.Text() + "\n"
				break
			}

		})
		lines = lines + line
	})

	utils.ExecuteWithRetry(func() error {
		return w.SendTextMessage(evt.GetThreadId(), lines)
	})

	return nil
}
