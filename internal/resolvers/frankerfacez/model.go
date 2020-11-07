package frankerfacez

import "time"

/* Example JSON data generated from https://api.frankerfacez.com/v1/emote/131001 2020-11-07
{
  "emote": {
    "created_at": "Sun, 25 Sep 2016 12:30:30 GMT",
    "css": null,
    "height": 21,
    "hidden": false,
    "id": 131001,
    "last_updated": "Sun, 25 Sep 2016 14:25:01 GMT",
    "margins": null,
    "modifier": false,
    "name": "pajaE",
    "offset": null,
    "owner": {
      "_id": 63119,
      "display_name": "pajaSWA",
      "name": "pajaswa"
    },
    "public": true,
    "status": 1,
    "urls": {
      "1": "//cdn.frankerfacez.com/8542ab940f02f3bdc938796dc7258902.PNG",
      "2": "//cdn.frankerfacez.com/c1fe2e20b1d13e97b40b44f6893a7ba4.PNG",
      "4": "//cdn.frankerfacez.com/6154d1c0f922ee6506cb2e555dd46e03.png"
    },
    "usage_count": 9,
    "width": 32
  }
}
*/
type FrankerFaceZEmoteAPIResponse struct {
	Height       int16  `json:"height"`
	Modifier     bool   `json:"modifier"`
	Status       int    `json:"status"`
	Width        int16  `json:"width"`
	Hidden       bool   `json:"hidden"`
	CreatedAtRaw string `json:"created_at"`
	CreatedAt    time.Time
	UpdatedAtRaw string `json:"last_updated"`
	UpdatedAt    time.Time
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Public       bool   `json:"public"`
	Owner        struct {
		DisplayName string `json:"display_name"`
		ID          int    `json:"_id"`
		Name        string `json:"name"`
	} `json:"owner"`

	URLs struct {
		Size1 string `json:"1"`
		Size2 string `json:"2"`
		Size4 string `json:"4"`
	} `json:"urls"`
}

type TooltipData struct {
	Code     string
	Uploader string
}
