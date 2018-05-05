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

func ProcessStringConsts(file *DomainFile) []ProcessedStringConst {
	consts := []ProcessedStringConst{}

	for _, def := range file.StringConsts {
		consts = append(consts, ProcessedStringConst{
			Key:   def.Key,
			Value: def.Value,
		})
	}

	return consts
}

func GenerateEnumsAndConsts(data *TplData) error {
	if err := WriteTemplateFile("../pkg/domain/domain.go", data, DomainFileTemplateGo); err != nil {
		return err
	}

	if err := WriteTemplateFile("../frontend/generated/domain.ts", data, DomainFileTemplateTypeScript); err != nil {
		return err
	}

	return nil
}
