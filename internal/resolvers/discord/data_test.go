package discord

import (
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

var (
	data_raw = map[string][]byte{}
)

func init() {
	data_raw["bad"] = []byte(`xD`)
	data_raw["forsen"] = []byte(`{"type":0,"code":"forsen","expires_at":null,"flags":2,"guild":{"id":"97034666673975296","name":"Forsen","splash":"05b8f7eb7f06f11da324945b0bac65ee","banner":"a_b10dd2b4e2c25b002ad9c303432a373c","description":null,"icon":"a_ea433153b6ce120e0fb518efc084dc38","features":["SEVEN_DAY_THREAD_ARCHIVE","MEMBER_PROFILES","PRIVATE_THREADS","ANIMATED_ICON","VANITY_URL","THREE_DAY_THREAD_ARCHIVE","ROLE_ICONS","AUTO_MODERATION","ANIMATED_BANNER","NEW_THREAD_PERMISSIONS","INVITE_SPLASH","THREADS_ENABLED","CHANNEL_ICON_EMOJIS_GENERATED","NON_COMMUNITY_RAID_ALERTS","BANNER","SOUNDBOARD"],"verification_level":3,"vanity_url_code":"forsen","nsfw_level":0,"nsfw":false,"premium_subscription_count":107},"guild_id":"97034666673975296","channel":{"id":"97034666673975296","type":0,"name":"readme"},"approximate_member_count":44960,"approximate_presence_count":13730}`)
	data_raw["qbRE8WR"] = []byte(`{"type":0,"code":"qbRE8WR","inviter":{"id":"85699361769553920","username":"pajlada","avatar":"e75df3dbe6cb04b3c9f0e090b3adb190","discriminator":"0","public_flags":512,"flags":512,"banner":null,"accent_color":13387007,"global_name":"pajlada","avatar_decoration_data":null,"banner_color":"#cc44ff","clan":null},"expires_at":null,"flags":2,"guild":{"id":"138009976613502976","name":"pajlada","splash":null,"banner":null,"description":null,"icon":"dcbac612ccdd3ffa2fbf89647e26f929","features":["CHANNEL_ICON_EMOJIS_GENERATED","INVITE_SPLASH","THREE_DAY_THREAD_ARCHIVE","COMMUNITY","ANIMATED_ICON","SOUNDBOARD","NEW_THREAD_PERMISSIONS","ACTIVITY_FEED_DISABLED_BY_USER","THREADS_ENABLED","NEWS"],"verification_level":1,"vanity_url_code":null,"nsfw_level":0,"nsfw":false,"premium_subscription_count":6},"guild_id":"138009976613502976","channel":{"id":"138009976613502976","type":0,"name":"general"},"approximate_member_count":1515,"approximate_presence_count":563}`)
}

func testServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/api/v9/invites/{invite}", func(w http.ResponseWriter, r *http.Request) {
		invite := chi.URLParam(r, "invite")

		w.Header().Set("Content-Type", "application/json")

		if response, ok := data_raw[invite]; ok {
			w.Write(response)
		} else {
			http.Error(w, http.StatusText(404), 404)
		}
	})
	return httptest.NewServer(r)
}
