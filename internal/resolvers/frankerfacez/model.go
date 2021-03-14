package frankerfacez

import "time"

/* Example JSON data generated from https://api.frankerfacez.com/v1/emote/131001 2020-11-18
{
  "emote": {
    "created_at": "2016-09-25T12:30:30.313Z",
    "css": null,
    "height": 21,
    "hidden": false,
    "id": 131001,
    "last_updated": "2016-09-25T14:25:01.408Z",
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
      "1": "//cdn.frankerfacez.com/emote/131001/1",
      "2": "//cdn.frankerfacez.com/emote/131001/2",
      "4": "//cdn.frankerfacez.com/emote/131001/4"
    },
    "usage_count": 9,
    "width": 32
  }
}
*/

type FrankerFaceZEmoteAPIResponse struct {
	Height    int16     `json:"height"`
	Modifier  bool      `json:"modifier"`
	Status    int       `json:"status"`
	Width     int16     `json:"width"`
	Hidden    bool      `json:"hidden"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"last_updated"`
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Public    bool      `json:"public"`
	Owner     struct {
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
