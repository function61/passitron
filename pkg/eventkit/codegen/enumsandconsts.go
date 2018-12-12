package codegen

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

var enumValueCamelCaser = regexp.MustCompile("[^a-zA-Z0-9]+")

// "abba_cd" => "AbbaCd"
func camelCaseEnumValue(in string) string {
	return strings.Replace(
		strings.Title(
			enumValueCamelCaser.ReplaceAllString(in, " ")),
		" ",
		"",
		-1)
}

func ProcessStringEnums(file *DomainFile) []ProcessedStringEnum {
	enums := []ProcessedStringEnum{}

	for _, enum := range file.Enums {
		if enum.Type != "string" {
			panic(errors.New("unknown enum type: " + enum.Type))
		}

		members := []ProcessedStringEnumMember{}

		membersDigest := sha1.Sum([]byte(strings.Join(enum.StringMembers, ",")))

		for _, member := range enum.StringMembers {
			members = append(members, ProcessedStringEnumMember{
				Key:     camelCaseEnumValue(member),
				GoKey:   enum.Name + camelCaseEnumValue(member),
				GoValue: member,
			})
		}

		enums = append(enums, ProcessedStringEnum{
			Name:          enum.Name,
			MembersDigest: hex.EncodeToString(membersDigest[:])[0:6],
			Members:       members,
		})
	}

	return enums
}
