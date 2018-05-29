package emoji

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func BuildTable(searchDir string) error {
	fileList := []string{}

	if err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() == false {
			fileList = append(fileList, path)
		}
		return nil
	}); err != nil {
		return err
	}

	for _, file := range fileList {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
			} else if line[0] == '#' {
			} else if strings.ContainsAny(line, "{}") {
				if lines, err := expandShortNameLine(line); err != nil {
					fmt.Println("err: ", err)
					continue
				} else {
					for i := range lines {
						if err := loadShortName(lines[i]); err != nil {
							fmt.Println(err.Error(), " in ", f.Name())
						}
					}
				}
			} else {
				if err := loadShortName(line); err != nil {
					fmt.Println(err.Error(), " in ", f.Name())
				}
			}
		}

		if err := sc.Err(); err != nil {
			return err
		}
	}

	fmt.Println("emojiMap size: ", len(emojiMap))

	return nil
}

func loadShortName(line string) error {
	parts := strings.Split(line, ";")
	if len(parts) != 2 {
		fmt.Errorf("%s format error", line)
	} else {
		keys := strings.Split(parts[1], "/")
		for _, k := range keys {
			key := ":" + k + ":"

			if _, ok := emojiMap[key]; !ok {
				unicodes := strings.Split(parts[0], "-")
				value := ""
				for _, s := range unicodes {
					value += fmt.Sprintf("\\U%08s", s)
				}

				emojiMap[key] = value
			} else {
				return fmt.Errorf("%s already existed", line)
			}
		}
	}

	return nil
}

var expandReplacements = []struct {
	re          *regexp.Regexp
	replacement []string
}{
	{regexp.MustCompile(`{GENDER}`), []string{"male", "female"}},
	{regexp.MustCompile(`{M/W}`), []string{"man", "woman"}},
	{regexp.MustCompile(`{MAN/WOMAN}`), []string{"1F468", "1F469"}},
	{regexp.MustCompile(`{MALE/FEMALE}`), []string{"2642-FE0F", "2640-FE0F"}},
	{regexp.MustCompile(`{SKIN}`), []string{"", "-1F3FB", "-1F3FC", "-1F3FD", "-1F3FE", "-1F3FF"}},
	{regexp.MustCompile(`{SKIN!}`), []string{"-1F3FB", "-1F3FC", "-1F3FD", "-1F3FE", "-1F3FF"}},
}

var blacklists = map[string][]string{
	"male":   []string{"1F469", "2640-FE0F"}, // male cannot be replaced with WOMAN and FEMALE
	"man":    []string{"1F469", "2640-FE0F"}, // man cannot be replaced with WOMAN and FEMALE
	"female": []string{"1F468", "2642-FE0F"}, // female cannot be replaced with MAN and MALE
	"woman":  []string{"1F468", "2642-FE0F"}, // woman cannot be replaced with MAN and MALE
}

func contains(list []string, s string) bool {
	for _, t := range list {
		if t == s {
			return true
		}
	}
	return false
}

func expandShortNameLine(line string, blacklist ...string) ([]string, error) {
	ret := []string{}

	for _, e := range expandReplacements {
		if e.re.MatchString(line) {
			for _, r := range e.replacement {
				if contains(blacklist, r) {
					continue
				}

				replacement := e.re.ReplaceAllString(line, r)
				if replacementRecusive, err := expandShortNameLine(replacement, blacklists[r]...); err != nil {
					fmt.Println("err ", err)
				} else {
					ret = append(ret, replacementRecusive...)
				}
			}
		}

	}

	if len(ret) == 0 {
		return []string{line}, nil
	}

	return ret, nil
}

func WriteToGo(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(`package emoji

var emojiMap = map[string]string{
`); err != nil {
		return err
	}

	for k, v := range emojiMap {
		str := fmt.Sprintf("\"%s\": \"%s\",\n", k, v)
		if _, err := f.WriteString(str); err != nil {
			return err
		}
	}

	if _, err := f.WriteString(`}`); err != nil {
		return err
	}

	f.Sync()
	exec.Command("gofmt " + filepath)
	return nil
}

// Emoji returns the unicode value for the given emoji. If the
// specified emoji does not exist, Emoji() returns the empty string.
func Emoji(emoji string) string {
	val, ok := emojiMap[emoji]
	if !ok {
		return emoji
	}
	return val
}

var reg = regexp.MustCompile("(:[\\w-]+:)")

// Emojitize takes in a string with emojis specified in it, and returns
// a string with every emoji place holder replaced with it's unicode value
// (unless it could not be found, in which case it is let alone).
func Emojitize(emojis string) string {
	return reg.ReplaceAllStringFunc(emojis, func(str string) string {
		return Emoji(str)
	})
}
