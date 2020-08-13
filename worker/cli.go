package worker

import (
	"bytes"
	"fmt"
	"github.com/ndphu/skypebot-go/media"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"path"
	"strings"
)

func (w *Worker) ConversationCommand() *cli.Command {
	return &cli.Command{
		Name:    "conversation",
		Aliases: []string{"cv"},
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l"},
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "limit return array",
						Value:   10,
					},
				},
				Action: func(c *cli.Context) error {
					limit := c.Int("limit")
					conversations, err := w.GetConversations(limit)
					if err != nil {
						return err
					}

					table := tablewriter.NewWriter(c.App.Writer)
					table.SetHeader([]string{"Name", "ID"})
					for _, conversation := range conversations {
						shortId := strings.TrimPrefix(conversation.Id, "19:")
						shortId = strings.TrimSuffix(shortId, "@thread.skype")
						table.Append([]string{shortId, conversation.Topic})
					}
					table.Render()
					return nil
				},
			},
			{
				Name:    "list-message",
				Aliases: []string{"lm"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "thread",
						Aliases:  []string{"t"},
						Usage:    "coversation ID to list messages",
						Required: true,
					},
				},
				Usage: "list all message in the conversation",
				Action: func(c *cli.Context) error {
					threadId := c.String("thread")
					if !strings.HasPrefix(threadId, "19:") {
						threadId = "19:" + threadId
					}
					if !strings.HasSuffix(threadId, "@thread.skype") {
						threadId = threadId + "@thread.skype"
					}
					messages, err := w.GetAllTextMessages(threadId, -1)
					if err != nil {
						fmt.Fprintln(c.App.Writer, "Fail to get message for thread:", threadId)
						fmt.Fprintln(c.App.Writer, "Error:", err.Error())
					} else {
						fmt.Fprintln(c.App.Writer, "Found:", len(messages), "messages")
					}
					return nil
				},
			},
			{
				Name:    "clear-message",
				Aliases: []string{"cm"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "thread",
						Aliases:  []string{"t"},
						Usage:    "coversation ID to remove",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					threadId := c.String("thread")
					if !strings.HasPrefix(threadId, "19:") {
						threadId = "19:" + threadId
					}
					if !strings.HasSuffix(threadId, "@thread.skype") {
						threadId = threadId + "@thread.skype"
					}
					messages, err := w.GetAllTextMessages(threadId, -1)
					if err != nil {
						fmt.Fprintln(c.App.Writer, "Fail to get message for thread:", threadId)
						fmt.Fprintln(c.App.Writer, "Error:", err.Error())
					} else {
						fmt.Fprintln(c.App.Writer, "Found:", len(messages), "messages")
						deletedCount := 0
						for _,msg := range messages {
							if path.Base(msg.From) == "8:" + w.skypeId && msg.SkypeEditedId == "" {
								if err := w.DeleteMessage(msg); err != nil {
									fmt.Fprintln(c.App.Writer, "Fail to delete message:", msg.Id)
								} else {
									deletedCount ++
								}
							}
						}
						fmt.Fprintln(c.App.Writer, "Deleted:", deletedCount, "messages")
					}
					return nil
				},
			},
		},
	}
}

func (w *Worker) CovidCommand() *cli.Command {
	return &cli.Command{
		Name: "covid19",
		Subcommands: []*cli.Command{
			{
				Name:    "update",
				Aliases: []string{"l"},
				Action: func(c *cli.Context) error {
					limit := c.Int("limit")
					conversations, err := w.GetConversations(limit)
					if err != nil {
						return err
					}

					table := tablewriter.NewWriter(c.App.Writer)
					table.SetHeader([]string{"Name", "ID"})
					for _, conversation := range conversations {
						table.Append([]string{conversation.Id, conversation.Topic})
					}
					table.Render()
					return nil
				},
			},
		},
	}
}

func (w *Worker) NsfwCommand(threadId string) *cli.Command {
	var commands []*cli.Command

	commands = append(commands, &cli.Command{
		Name: "keyword",
		Action: func(c *cli.Context) error {
			fmt.Fprintln(c.App.Writer, "Available keyword:")
			fmt.Fprintln(c.App.Writer, strings.Join(media.GetKeywords(), ", "))
			return nil
		},
	})
	for _, kw := range media.GetKeywords() {
		x := kw
		commands = append(commands, &cli.Command{
			Name: x,
			Action: func(c *cli.Context) error {
				utils.ExecuteWithRetry(func() error {
					return w.sendRandomImage(threadId, x)
				})
				return nil
			},
		})
	}

	return &cli.Command{
		Name: "nsfw",
		Flags: []cli.Flag{

		},
		Subcommands: commands,
	}
}

func (w *Worker) NewAdminCLI(threadId string) *cli.App {
	return &cli.App{
		Name:      "admin-cli",
		UsageText: "skype-bot admin cli",

		Commands: []*cli.Command{
			w.ConversationCommand(),
			w.NsfwCommand(threadId),
		},

		ExitErrHandler: func(c *cli.Context, err error) {
		},
		CommandNotFound: func(c *cli.Context, cmd string) {
			if c.Command.Name == "nsfw" {
				//TODO hack
				utils.ExecuteWithRetry(func() error {
					return w.sendRandomImage(threadId, cmd)
				})
			} else {
				fmt.Fprintf(c.App.Writer, "Command not found: [%s]. See help message for details.\n", cmd)
			}
		},
	}
}

func (w *Worker) HandleAdminCommand(event *model.MessageEvent) error {
	adminCLI := w.NewAdminCLI(event.GetThreadId())
	var buff bytes.Buffer
	adminCLI.Writer = &buff
	adminCLI.ErrWriter = &buff
	content := normalizeMessageContent("admin-cli " + event.Resource.Content)
	if err := adminCLI.Run(strings.Split(content, " ")); err != nil {

		return w.SendTextMessage(event.GetThreadId(), "<pre>"+buff.String()+"</pre>")
	}
	return utils.ExecuteWithRetry(func() error {
		if buff.Len() == 0 {
			return nil
		}
		return w.SendTextMessage(event.GetThreadId(), "<pre>"+buff.String()+"</pre>")
	})
}
