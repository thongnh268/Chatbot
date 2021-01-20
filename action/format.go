package action

import (
	"encoding/json"
	"strconv"

	"github.com/subiz/goutils/log"
	pb "github.com/subiz/header/common"
)

type FormatText struct {
	Text string
	Key  string
}

func unmarshalQuillDelta(str string) []*FormatText {
	out := make([]*FormatText, 0)
	var rooti interface{}
	err := json.Unmarshal([]byte(str), &rooti)
	if err != nil {
		return out
	}
	root, ok := rooti.(map[string]interface{})
	var formatis []interface{}
	if ok && root["ops"] != nil {
		formatis = root["ops"].([]interface{})
	} else {
		json.Unmarshal([]byte(str), &formatis) // TODO remove
	}
	if formatis == nil || len(formatis) == 0 {
		return out
	}
	for _, formati := range formatis {
		format, ok := formati.(map[string]interface{})
		if !ok || format["insert"] == nil {
			continue
		}
		text, ok := format["insert"].(string)
		if ok {
			out = append(out, &FormatText{Text: text})
			continue
		}
		insert, ok := format["insert"].(map[string]interface{})
		if !ok || insert["dynamicField"] == nil {
			continue
		}
		dynamicf, ok := insert["dynamicField"].(map[string]interface{})
		if !ok || dynamicf["key"] == nil {
			continue
		}
		key, ok := dynamicf["key"].(string)
		if ok && key != "" {
			out = append(out, &FormatText{Key: key})
		}
	}
	return out
}

type ObjectTree struct {
	data map[string]interface{}
}

func (tree *ObjectTree) GetKVs(path string, keys []string) []*pb.KV {
	out := make([]*pb.KV, 0)
	if path == "" {
		return out
	}
	pathParts := Regex_attr_name.FindAllString(path, -1)
	if len(pathParts) == 0 {
		return out
	}
	pathi := parseTree(tree.data, pathParts)
	pathiarr, ok := pathi.([]interface{})
	if !ok || len(pathiarr) == 0 {
		return out
	}
	for _, rowi := range pathiarr {
		for _, key := range keys {
			if key == "" {
				continue
			}
			keyPaths := Regex_attr_name.FindAllString(key, -1)
			if len(keyPaths) == 0 {
				continue
			}
			texti := parseTree(rowi, keyPaths)
			text, ok := texti.(string)
			if ok && text != "" {
				out = append(out, &pb.KV{Key: key, Value: text})
				break
			}
		}
	}

	return out
}

func (tree *ObjectTree) GetText(key string) string {
	if key == "" {
		return ""
	}
	keyParts := Regex_attr_name.FindAllString(key, -1)
	if len(keyParts) == 0 {
		return ""
	}
	rootKey := keyParts[0]
	// TODO fetch root
	if _, has := tree.data[rootKey]; !has {
		return ""
	}
	texti := parseTree(tree.data, keyParts)
	text, ok := texti.(string)
	if ok {
		return text
	}
	return ""
}

func parseTree(treeData interface{}, keyParts []string) interface{} {
	var iter interface{}
	iter = treeData
	deep := 0
	index := 0
	for {
		if deep >= 64 {
			log.Warn("TOODEEPPP")
			return ""
		}
		if index >= len(keyParts) {
			break
		}
		if iter == nil {
			return ""
		}

		var nextiter interface{}
		itermap, ok := iter.(map[string]interface{})
		if ok && itermap[keyParts[index]] != nil {
			nextiter = itermap[keyParts[index]]
		}

		intKey, err := strconv.Atoi(keyParts[index])
		if err == nil && nextiter == nil {
			iterarr, ok := iter.([]interface{})
			if ok && intKey >= 0 && intKey < len(iterarr) {
				nextiter = iterarr[intKey]
			}
		}

		iter = nextiter
		index++
		deep++
	}

	return iter
}
