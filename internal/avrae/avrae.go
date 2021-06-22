package avrae

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rleszilm/genms/log"
)

type RollType int
type AttackType int

const (
	patterns string = `^\*\*(?P<initer>[a-zA-Z\s]*): Initiative: Roll\*\*: (?P<initDice>[0-9dklhproamie,\s+\-\(\)\*~]*)$|` +
		`^\*\*Total\*\*: (?P<roll>[0-9]+)$|` +
		`^\*\*To Hit\*\*: (?P<hitDice>[0-9dklhproamie,\s+\-\(\)\*~]*) = ` + "`" + `(?P<hit>[0-9]*)` + "`" + `$|` +
		`^\*\*Damage(?P<crit> \(CRIT!\))?\*\*: (?P<dmgDice>[0-9dklhproamie,\s+\-\(\)\*~]*) = ` + "`" + `(?P<dmg>[0-9]*)` + "`" + `$|` +
		`^(?P<miscDice>[0-9dklhproamie,\s+\-\(\)\*~]*) = ` + "`" + `(?P<misc>[0-9]*)` + "`" + `$|` +
		`^(?P<checkDesc>(?P<checker>[a-zA-Z\s]*) makes an? (?P<check>[a-zA-Z\s]*) check!)$|` +
		`^(?P<saveDesc>(?P<saver>[a-zA-Z\s]*) makes an? (?P<save>[a-zA-Z]*) Save!)$|` +
		`^(?P<attackDesc>(?P<attacker>[a-zA-Z\s]*) attacks with an? (?P<attack>[a-zA-Z\s]*)!)$|` +
		`^(?P<castDesc>(?P<caster>[a-zA-Z\s]*) casts (?P<spell>[a-zA-Z\s]*)!)$|` +
		`^(?P<healDesc>(?P<healer>[a-zA-Z\s]*) heals with (?P<healSpell>[a-zA-Z\s]*)!)$`

	dieMinMaxPattern string = `\*\*([0-9]*)\*\*`
	dieDropPattern   string = `~~([0-9]*)~~`
)

var (
	parser      = regexp.MustCompile(patterns)
	dieMinMaxer = regexp.MustCompile(dieMinMaxPattern)
	dieDropper  = regexp.MustCompile(dieDropPattern)

	logs = log.NewChannel("avrae")
)

type Roll struct {
	ID              string    `json:"id,omitempty"`
	Timestamp       time.Time `json:"created,omitempty"`
	EditedTimestamp time.Time `json:"edited,omitempty"`
	Description     string    `json:"description,omitempty"`
	Color           string    `json:"color,omitempty"`
	Player          string    `json:"player,omitempty"`
	Avatar          string    `json:"avatar,omitempty"`
	Roll            string    `json:"roll,omitempty"`
	Kind            string    `json:"kind,omitempty"`
	Attack          string    `json:"attack,omitempty"`
	Dice            string    `json:"dice,omitempty"`
	Total           string    `json:"total,omitempty"`
	HitDice         string    `json:"hitDice,omitempty"`
	HitTotal        string    `json:"hitTotal,omitempty"`
	DamageDice      string    `json:"damageDice,omitempty"`
	DamageTotal     string    `json:"damageTotal,omitempty"`
	Critical        bool      `json:"critical,omitempty"`
	Update          bool      `json:"update,omitempty"`
}

func NewCreateHandler(ch chan<- *Roll) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Author.String() != "Avrae#6944" {
			return
		}

		if roll := rollFromMessage(m.Message); roll != nil {
			ch <- roll
		}
	}
}

func NewUpdateHandler(ch chan<- *Roll) func(s *discordgo.Session, m *discordgo.MessageUpdate) {
	return func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Author.String() != "Avrae#6944" {
			return
		}

		if roll := rollFromMessage(m.Message); roll != nil {
			ch <- roll
		}
	}
}

func rollFromMessage(m *discordgo.Message) *Roll {
	var avatar string
	var color int
	tokens := map[string]string{}
	parse(m.Content, tokens)
	for _, e := range m.Embeds {
		parse(e.Description, tokens)
		parse(e.Title, tokens)

		for _, f := range e.Fields {
			parse(f.Value, tokens)
		}

		if e.Color != 0 {
			color = e.Color
		}

		if e.Thumbnail != nil {
			if e.Thumbnail.URL != "" {
				avatar = e.Thumbnail.URL
			}
		}
	}

	roll := rollFromTokens(tokens)
	roll.ID = m.ID
	roll.Avatar = avatar
	roll.Color = fmt.Sprintf("#%06x", color)

	t, err := m.Timestamp.Parse()
	if err != nil {
		logs.Warning(err)
	}
	roll.Timestamp = t

	e, err := m.EditedTimestamp.Parse()
	if err != nil {
		logs.Warning(err)
	}
	roll.EditedTimestamp = e

	return roll
}

func rollFromTokens(tokens map[string]string) *Roll {
	roll := &Roll{}

	if initer, ok := tokens["initer"]; ok {
		roll.Player = initer
		roll.Roll = "initiative"
		roll.Description = initer + " rolls for initiative!"
		roll.Kind = "Initiative"
		roll.Dice = tokens["initDice"]
		roll.Total = tokens["roll"]
	} else if checker, ok := tokens["checker"]; ok {
		roll.Player = checker
		roll.Roll = "check"
		roll.Description = tokens["checkDesc"]
		roll.Kind = tokens["check"]
		roll.Dice = tokens["miscDice"]
		roll.Total = tokens["misc"]
	} else if saver, ok := tokens["saver"]; ok {
		roll.Player = saver
		roll.Roll = "save"
		roll.Description = tokens["saveDesc"]
		roll.Kind = tokens["save"]
		roll.Dice = tokens["miscDice"]
		roll.Total = tokens["misc"]
	} else if attacker, ok := tokens["attacker"]; ok {
		roll.Player = attacker
		roll.Roll = "action"
		roll.Description = tokens["attackDesc"]
		roll.Attack = "melee"
		roll.Kind = tokens["attack"]
		roll.HitDice = tokens["hitDice"]
		roll.HitTotal = tokens["hit"]
		roll.DamageDice = tokens["dmgDice"]
		roll.DamageTotal = tokens["dmg"]

		if _, ok := tokens["crit"]; ok {
			roll.Critical = true
		}
	} else if caster, ok := tokens["caster"]; ok {
		roll.Player = caster
		roll.Roll = "action"
		roll.Description = tokens["castDesc"]
		roll.Attack = "magical"
		roll.Kind = tokens["spell"]
		roll.HitDice = tokens["hitDice"]
		roll.HitTotal = tokens["hit"]
		roll.DamageDice = tokens["dmgDice"]
		roll.DamageTotal = tokens["dmg"]

		if _, ok := tokens["crit"]; ok {
			roll.Critical = true
		}
	} else if healer, ok := tokens["healer"]; ok {
		roll.Player = healer
		roll.Roll = "action"
		roll.Description = tokens["healDesc"]
		roll.Attack = "healing"
		roll.Kind = tokens["healSpell"]
		roll.DamageDice = tokens["miscDice"]
		roll.DamageTotal = tokens["misc"]
	}

	roll.Dice = dieMinMaxer.ReplaceAllStringFunc(roll.Dice, dieMinMax)
	roll.DamageDice = dieMinMaxer.ReplaceAllStringFunc(roll.DamageDice, dieMinMax)
	roll.HitDice = dieMinMaxer.ReplaceAllStringFunc(roll.HitDice, dieMinMax)

	roll.Dice = dieDropper.ReplaceAllStringFunc(roll.Dice, dieDrop)
	roll.DamageDice = dieDropper.ReplaceAllStringFunc(roll.DamageDice, dieDrop)
	roll.HitDice = dieDropper.ReplaceAllStringFunc(roll.HitDice, dieDrop)

	return roll
}

func dieMinMax(str string) string {
	return fmt.Sprintf(`<span class="die-minmaxed">%s</span>`, str[2:len(str)-2])
}

func dieDrop(str string) string {
	return fmt.Sprintf(`<span class="die-dropped">%s</span>`, str[2:len(str)-2])
}

func parse(str string, tokens map[string]string) {
	strs := strings.Split(str, "\n")
	for _, s := range strs {
		matches := parser.FindAllStringSubmatch(s, -1)

		if matches == nil {
			continue
		}

		for _, match := range matches {
			for i, name := range parser.SubexpNames() {
				if name != "" && match[i] != "" {
					tokens[name] = match[i]
				}
			}
		}
	}
}
