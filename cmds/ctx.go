package cmds

import (
	"errors"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/TeamWhiskey/whiskey/args"
	"github.com/TeamWhiskey/whiskey/utils"
	"github.com/TeamWhiskey/whiskey/utils/msgcol"
	"github.com/bwmarrin/discordgo"
)

const (
	// PromptAccept for acceptance of a prompt.
	PromptAccept = iota

	// PromptDeny for denial of a prompt.
	PromptDeny

	// PromptTimeout is when the prompt times out.
	PromptTimeout
)

// Context is the message context used for command execution
type Context struct {
	Session *discordgo.Session
	Msg     *discordgo.Message
	Cmd     *Command

	lastMsg *discordgo.Message

	// String used to call the command
	CmdCallKey string
	ArgsCtx    *args.ArgumentsContext

	StartTime time.Time
}

// Send sends a message to the channel the msg came from.
func (ctx *Context) Send(content string, deleteTime ...time.Duration) (*discordgo.Message, error) {
	m, err := ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, content)
	ctx.lastMsg = m

	if len(deleteTime) > 0 {
		go (func() {
			time.Sleep(deleteTime[0])
			ctx.Delete(m)
		})()
	}

	return m, err
}

// Edit edits the last msg sent by the context.
func (ctx *Context) Edit(content string) (*discordgo.Message, error) {
	if ctx.lastMsg == nil {
		return nil, errors.New("Tried to edit a message that does not exist")
	}

	m, err := ctx.Session.ChannelMessageEdit(ctx.lastMsg.ChannelID, ctx.lastMsg.ID, content)
	ctx.lastMsg = m
	return m, err
}

// Delete deletes the last recently sent message or the supplied one.
func (ctx *Context) Delete(msg ...*discordgo.Message) error {
	msgSupplied := len(msg) > 0
	if msgSupplied && ctx.lastMsg == nil {
		return errors.New("Tried to delete a message that does not exist")
	}

	var ID string
	var chanID string
	if msgSupplied {
		ID = msg[0].ID
		chanID = msg[0].ChannelID
	} else {
		ID = ctx.lastMsg.ID
		chanID = ctx.lastMsg.ChannelID
	}

	return ctx.Session.ChannelMessageDelete(chanID, ID)
}

// Prompt sends a prompt to accept or deny within 10 seconds.
func (ctx *Context) Prompt(prompt string) (int, error) {
	_, err := ctx.Send(prompt + " This action will cancel in 10 seconds. [y/N]")
	if err != nil {
		return PromptDeny, err
	}

	msgs, err := ctx.Collect(10 * time.Second)
	if err != nil {
		return PromptDeny, err
	}

	accepted := false
	denied := false
	for msg := range msgs {
		switch strings.ToLower(msg.Content) {
		case "y":
			accepted = true
			break
		case "n":
			denied = true
			break
		}
	}

	if err = ctx.Delete(); err != nil {
		return PromptDeny, err
	}

	if !accepted {
		if denied {
			ctx.Send("Cancelling purge.")

			return PromptDeny, nil
		}

		ctx.Send("Timed out.")
		return PromptTimeout, nil
	}

	return PromptAccept, nil
}

// CleanUp deletes the last recently sent message.
func (ctx *Context) CleanUp() {
	if ctx.lastMsg != nil {
		time.Sleep(10 * time.Second)
		ctx.Session.ChannelMessageDelete(ctx.lastMsg.ChannelID, ctx.lastMsg.ID)
	}
}

// Collect collects messages.
func (ctx *Context) Collect(time time.Duration, amnt ...int) (chan *discordgo.Message, error) {
	amntEmpty := len(amnt) == 0
	if time == 0 && amntEmpty {
		return nil, errors.New("ctx.Collect can not collect indefinitely if no amount is supplied")
	}

	var amntToPurge int
	if amntEmpty {
		amntToPurge = 0
	} else {
		amntToPurge = amnt[0]
	}

	colMnger := msgcol.GetCollectionManager()
	col := colMnger.NewCollector(ctx.Msg.ChannelID, time, amntToPurge)
	return col.Msgs, nil
}

// SendError reports an error to the err channel and to the user
func (ctx *Context) SendError(err error) {
	cmdName := (func() string {
		if ctx.Cmd != nil && ctx.Cmd.Name != "" {
			return ctx.Cmd.Name
		}

		return "N/A"
	})()

	errTxt := ":rotating_light: An error occurred while handling the command `" + cmdName + "`:\n```" + err.Error() + "```"
	ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, errTxt+"\nThe error has been reported")
	config := utils.GetConfig()
	if config.LogChannel != "" {
		stacktrace := string(debug.Stack())
		log.Print("An error occurred while handling the command " + cmdName + ":\n" + err.Error() + "\n" + stacktrace)
		ctx.Session.ChannelMessageSend(config.LogChannel, errTxt+"\n\n**Stacktrace**\n```"+stacktrace+"```")
	}
}
