package commands

func (h *handler) handleGhosts() error {
	return h.say("To submit ghost alerts, cheer 200 bits and include 'ghost of <whatever>' in your message. To use 200 fun points from your balance, send '!ghost of <whatever>' as a normal message.")
}

func (h *handler) handleFriends() error {
	return h.say("To submit friend alerts, cheer 200 bits and include 'friend <whatever>' in your message. To use 200 fun points from your balance, send '!friend <whatever>' as a normal message.")
}

func (h *handler) handleAlerts() error {
	return h.say("You can cheer 200 bits and mention prayer bear, or you can cheer 300 bits and ask us to stand back. !prayerbear and !standback also work if you have the fun points to spend.")
}

func (h *handler) handleTapes() error {
	return h.say("Browse tapes at https://goldenvcr.com/tapes - you can log in with Twitch and mark tapes you want to see as favorites.")
}

func (h *handler) handleRemix() error {
	return h.say("Cheers for 1000 bits are honored as song requests. Choose from any of these clips: https://goldenvcr.com/remix")
}

func (h *handler) handleYoutube() error {
	return h.say("Watch VODs and clips on YouTube: https://www.youtube.com/@GoldenVCR/videos")
}

func (h *handler) handleCamera() error {
	return h.say("A camera is a device for recording visual images in the form of photographs, film, or video signals.")
}
