package commands

func (h *handler) handleGhosts() error {
	return h.say("To submit ghost alerts, cheer 200 bits and include 'ghost of <whatever>' in your message. To use your existing balance of Fun Points, prefix your message with '!200' instead of cheering.")
}

func (h *handler) handleTapes() error {
	return h.say("Browse tapes at https://goldenvcr.com/tapes - you can log in with Twitch and mark tapes you want to see as favorites.")
}

func (h *handler) handleYoutube() error {
	return h.say("Watch VODs and clips on YouTube: https://www.youtube.com/@GoldenVCR/videos")
}
