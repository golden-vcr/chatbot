package commands

func (h *handler) handleGhosts() error {
	return h.say("To submit ghost alerts, either: 1.) cheer 200 bits with a message containing \"ghost of <thing you want to see>\", or 2.) log in to goldenvcr.com and use the form on the front page to spend your Golden VCR Fun Points.")
}

func (h *handler) handleTapes() error {
	return h.say("Browse tapes at https://goldenvcr.com/tapes - you can log in with Twitch and mark tapes you want to see as favorites.")
}

func (h *handler) handleYoutube() error {
	return h.say("Watch VODs and clips on YouTube: https://www.youtube.com/@GoldenVCR/videos")
}
