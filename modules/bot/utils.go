package bot

import "gopkg.in/tucnak/telebot.v2"

// GetURLAndMentionFromMessage get URL and mention from message
func GetURLAndMentionFromMessage(m *telebot.Message) (url string, mention string) {
	for _, entity := range m.Entities {
		if entity.Type == telebot.EntityMention {
			if mention == "" {
				mention = m.Text[entity.Offset : entity.Offset+entity.Length]

			}
		}

		if entity.Type == telebot.EntityURL {
			if url == "" {
				url = m.Text[entity.Offset : entity.Offset+entity.Length]
			}
		}
	}

	return
}
