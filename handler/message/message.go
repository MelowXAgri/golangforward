package message

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tgbot/config"
	"github.com/amarnathcjd/gogram/telegram"
	//"github.com/davecgh/go-spew/spew"
)

var fwCount int = 50

func RegisterMessageHandler(bot *telegram.Client, user *telegram.Client, cfg *config.Config) {
	notChannel := telegram.FilterFunc(func(m *telegram.NewMessage) bool {
		return m.ChatType() != "channel"
	})
	isChannel := telegram.FilterFunc(func(m *telegram.NewMessage) bool {
		return m.ChatType() == "channel"
	})
	isAdmin := telegram.FilterFunc(func(m *telegram.NewMessage) bool {
		for _, id := range cfg.Admin {
			if id == m.SenderID() {
				return true
			}
		}
		return false
	})

	// /start command
	bot.On("message:/start", func(msg *telegram.NewMessage) error {
		ret := "/channelsrc [channelId] -> Channel Source\n"
		ret += "/channeldst [channelId] -> Channel destination\n"
		ret += "/autostart [on|off]\n"
		ret += "/fw -> Start forward\n"
		ret += "/getid -> Get id of channel, group or your."
		msg.Reply(ret)
		return nil
	}, notChannel, isAdmin)

	bot.On("message:/autostart", func(msg *telegram.NewMessage) error {
		args := strings.Fields(msg.Text())
		if len(args) < 2 {
			msg.Reply("❌ Usage: /autostart [on|off]")
			return nil
		}
		raw := args[1]
		if strings.ToLower(raw) == "on" {
			cfg.AutoFetch = true
			msg.Reply("Auto forward new videos active.")
		} else if strings.ToLower(raw) == "off" {
			msg.Reply("Auto forward new videos disable.")
			cfg.AutoFetch = false
		} else {
			msg.Reply("❌ Usage: /autostart [on|off]")
		}
		cfg.Save()
		return nil
	}, isAdmin)

	// /channelsrc command
	bot.On("message:/channelsrc", func(msg *telegram.NewMessage) error {
		args := strings.Fields(msg.Text())
		if len(args) < 2 {
			msg.Reply("❌ Usage: /channelsrc [channelID]")
			return nil
		}

		raw := args[1]
		if strings.HasPrefix(raw, "-100") {
			raw = raw[4:]
		}

		channelID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			msg.Reply("❌ Invalid channel ID")
			return err
		}

		ch, err := bot.GetChannel(channelID)
		if err != nil {
			msg.Reply(fmt.Sprintf("❌ Failed to get channel: %v", err))
			return err
		}

		msg.Reply(fmt.Sprintf("✅ Channel info:\nID: %v\nTitle: %v", ch.ID, ch.Title))
		cfg.ChannelSrc = channelID
		cfg.Save()
		return nil
	}, notChannel, isAdmin)

	// /channeldst command
	bot.On("message:/channeldst", func(msg *telegram.NewMessage) error {
		args := strings.Fields(msg.Text())
		if len(args) < 2 {
			msg.Reply("❌ Usage: /channeldst [channelID]")
			return nil
		}
		raw := args[1]
		if strings.HasPrefix(raw, "-100") {
			raw = raw[4:]
		}

		channelID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			msg.Reply("❌ Invalid channel ID")
			return err
		}

		ch, err := bot.GetChannel(channelID)
		if err != nil {
			msg.Reply(fmt.Sprintf("❌ Failed to get channel: %v", err))
			return err
		}

		msg.Reply(fmt.Sprintf("✅ Channel info:\nID: %v\nTitle: %v", ch.ID, ch.Title))
		cfg.ChannelDst = channelID
		cfg.Save()
		return nil
	}, notChannel, isAdmin)

	// /getid command
	bot.On("message:/getid", func(msg *telegram.NewMessage) error {
		switch msg.ChatType() {
		case "channel":
			peer := &telegram.InputPeerChannel{
				ChannelID:  msg.Channel.ID,
				AccessHash: msg.Channel.AccessHash,
			}
			_, _ = bot.SendMessage(peer, fmt.Sprintf("ChannelID: %v", msg.Channel.ID), nil)

		case "chat":
			msg.Reply(fmt.Sprintf("GroupID: %v", msg.ChannelID()))

		default:
			msg.Reply(fmt.Sprintf("Your UserID: %v", msg.Sender.ID))
		}
		return nil
	})

	bot.On("message:/fw", func(msg *telegram.NewMessage) error {
		var anu *telegram.NewMessage
		if cfg.OnStart {
			msg.Reply("Oops, Still in progress.")
			return nil
		}
		if cfg.ChannelSrc == 0 {
			msg.Reply("Oops channel source is undefined.")
			return nil
		}
		srcCh, err := user.GetChannel(cfg.ChannelSrc)
		if err != nil {
			msg.Reply("Oops channel not found!")
			return err
		}
		if srcCh.Noforwards {
			msg.Reply("Oops channel Restricted.")
			return nil
		}

		if !IsAdmin(user, srcCh.ID, srcCh.AccessHash) {
			msg.Reply("Oops, im not admin on that Channel.")
			return nil
		}
		
		srcPeer := &telegram.InputPeerChannel{
			ChannelID:  srcCh.ID,
			AccessHash: srcCh.AccessHash,
		}

		anu, _ = msg.Reply("Please wait, fetching channel source ~...")
		time.Sleep(2 * time.Second)
		msgs, err := GetAllHistory(user, srcPeer, 0, 0) // 0 ambil semua
		if err != nil {
			fmt.Println(err)
			return err
		}
		if len(msgs) == 0 {
			anu.Edit("No new videos to fetch...", telegram.SendOptions{})
			return nil
		}
		time.Sleep(2 * time.Second)
		
		if cfg.ChannelDst == 0 {
			msg.Reply("Oops channel destination is undefined.")
			return nil
		}
		anu, _ = anu.Edit("Channel fetched, filtering videos...", telegram.SendOptions{})
		dstCh, err := bot.GetChannel(cfg.ChannelDst)
		if err != nil {
			fmt.Println("GetChannel error:", err)
			return err
		}
		if !IsAdmin(bot, dstCh.ID, dstCh.AccessHash) {
			msg.Reply("Oops, lol im not admin on that Channel.")
			return nil
		}
		srcCh, err = bot.GetChannel(cfg.ChannelSrc)
		if err != nil {
			fmt.Println("GetChannel error:", err)
			return err
		}
		srcPeer = &telegram.InputPeerChannel{
			ChannelID:  srcCh.ID,
			AccessHash: srcCh.AccessHash,
		}

		dstPeer := &telegram.InputPeerChannel{
			ChannelID:  dstCh.ID,
			AccessHash: dstCh.AccessHash,
		}

		var docs []telegram.NewMessage
		for _, m := range msgs {
			if m.MediaType() == "document" {
				docs = append(docs, m)
			}
		}
		count := len(docs)
		time.Sleep(3 * time.Second)
		if len(docs) == 0 {
			anu, _ = anu.Edit(fmt.Sprintf("Total: %d videos fetched. Nothing to send...", len(docs)), telegram.SendOptions{})
			return nil
		} else {
			anu, _ = anu.Edit(fmt.Sprintf("Total: %d videos fetched. Sending videos...", len(docs)), telegram.SendOptions{})
		}
		seenID := make(map[int32]bool)
		for i := 0; i < len(docs); i += fwCount {
            var msgIds []int32
			for _, m := range docs[i:min(i+fwCount, len(docs))] {
				if !seenID[int32(m.ID)] {
					seenID[int32(m.ID)] = true
					msgIds = append(msgIds, int32(m.ID))
				}
			}
			fmt.Printf("[FORWARD] Batch %d/%d, forwarding (%d/%d) messages...\n", (i/fwCount)+1, (len(docs)+fwCount-1)/fwCount, len(msgIds), count)
			opts := &telegram.ForwardOptions{HideAuthor: true}
			if err := safeForward(bot, dstPeer, srcPeer, msgIds, opts); err != nil {
				fmt.Println(err)
			}
			count -= len(msgIds)
			msgIds = []int32{}
			time.Sleep(5 * time.Second)
		}
		anu.Edit("All video successfully send...", telegram.SendOptions{})

		return nil
	}, notChannel, isAdmin)

	// user auto fetch
	bot.On(telegram.OnMessage, func(msg *telegram.NewMessage) error {
		if !cfg.AutoFetch {
			return nil
		}
		if cfg.ChannelDst == 0 {
			fmt.Println("Oops channel destination is undefined.")
			return nil
		}
		dstCh, err := bot.GetChannel(cfg.ChannelDst)
		if err != nil {
			fmt.Println("GetChannel error:", err)
			return err
		}
		if !IsAdmin(bot, dstCh.ID, dstCh.AccessHash) {
			fmt.Println("Oops, lol im not admin on that Channel Destination")
			return nil
		}
		srcCh, err := bot.GetChannel(cfg.ChannelSrc)
		if err != nil {
			fmt.Println("GetChannel error:", err)
			return err
		}
		srcPeer := &telegram.InputPeerChannel{
			ChannelID:  srcCh.ID,
			AccessHash: srcCh.AccessHash,
		}
		dstPeer := &telegram.InputPeerChannel{
			ChannelID:  dstCh.ID,
			AccessHash: dstCh.AccessHash,
		}
		if msg.MediaType() == "document" {
			opts := &telegram.ForwardOptions{HideAuthor: true}
			if err := safeForward(bot, dstPeer, srcPeer, []int32{msg.ID}, opts); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}, isChannel)
}

// GetAllHistory fetches messages until there are no more.
// peerID should implement telegram.InputPeer (eg. *telegram.InputPeerChannel).
func GetAllHistory(c *telegram.Client, peerID telegram.InputPeer, totalLimit, offset int32) ([]telegram.NewMessage, error) {
	var allMessages []telegram.NewMessage
	batch := 0
	start := time.Now()

	for {
		remaining := totalLimit - int32(len(allMessages))
		if totalLimit > 0 && remaining <= 0 {
			break
		}

		limit := int32(100)
		if remaining > 0 && remaining < limit {
			limit = remaining
		}

		batch++
		fmt.Printf("[FETCH] Batch %d start, offset=%d, limit=%d\n", batch, offset, limit)
		opt := &telegram.HistoryOption{
			Limit:            limit,
			Offset:           offset,
			SleepThresholdMs: 500,
		}
		msgs, err := c.GetHistory(peerID, opt)
		if err != nil {
			fmt.Printf("[FETCH] Batch %d error: %v\n", batch, err)
			return allMessages, err
		}

		if len(msgs) == 0 {
			fmt.Printf("[FETCH] Batch %d done, no more messages. Total fetched=%d, elapsed=%v\n",
				batch, len(allMessages), time.Since(start))
			break
		}

		allMessages = append(allMessages, msgs...)
		// set offset to the last message id - 1 to avoid refetching the same message
		offset = msgs[len(msgs)-1].ID

		fmt.Printf("[FETCH] Batch %d ok, got %d messages, total=%d, elapsed=%v\n",
			batch, len(msgs), len(allMessages), time.Since(start))

		// if totalLimit <= 0 we want to keep fetching until msgs == 0
		// loop continues
	}

	return allMessages, nil
}

// safeForward forwards messages from `fromPeer` to `dstPeer`, handling FLOOD_WAIT and worker busy errors.
func safeForward(bot *telegram.Client, dstPeer telegram.InputPeer, fromPeer telegram.InputPeer, msgIDs []int32, opts *telegram.ForwardOptions) error {
	for {
		if len(msgIDs) == 0 {
			return nil
		}
		_, err := bot.Forward(dstPeer, fromPeer, msgIDs, opts)
		if err != nil {
			if strings.Contains(err.Error(), "FLOOD_WAIT") {
				re := regexp.MustCompile(`(\d+)`)
				wait, _ := strconv.Atoi(re.FindString(err.Error()))
				fmt.Printf("[FORWARD] FLOOD WAIT %d seconds.\n", wait)
				time.Sleep(time.Duration(wait) * time.Second)
				continue
			} else if strings.Contains(err.Error(), "WORKER_BUSY_TOO_LONG_RETRY") {
				fmt.Printf("[FORWARD] WORKER BUSY, WAIT 5 seconds.\n")
				time.Sleep(5 * time.Second)
				if fwCount > 5 {
					fwCount -= 10
                    if fwCount <= 0 {
                        fwCount = 1
                    }
				} else if fwCount == 10 {
					fwCount = 5
				} else if fwCount == 5 {
					fwCount = 1
				} else if fwCount <= 0 {
					fwCount = 1
				}
                time.Sleep(5 * time)
                return nil
			}
			fmt.Println("Error forwarding:", err)
			return err
		}
		return nil
	}
}

func RemoveID(target int32, ids []int32) []int32 {
	result := make([]int32, 0, len(ids))
	for _, id := range ids {
		if id != target {
			result = append(result, id)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func IsAdmin(cl *telegram.Client, channelID int64, accessHash int64) bool {
	ch := &telegram.InputChannelObj{
		ChannelID:  channelID,
		AccessHash: accessHash,
	}

	self := &telegram.InputPeerSelf{}

	resp, err := cl.ChannelsGetParticipant(ch, self)
	if err != nil {
		fmt.Println(err, "error getParticipant.")
		return false
	}

	switch resp.Participant.(type) {
	case *telegram.ChannelParticipantCreator:
		return true
	case *telegram.ChannelParticipantAdmin:
		return true
	default:
		return false
	}
}
