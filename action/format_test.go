package action

import (
	"encoding/json"
	"reflect"
	"testing"

	pb "github.com/subiz/header/common"
	"github.com/tidwall/gjson"
)

// basic case, exception case, fixed case
// input, expected output, actual output

func TestUnmarshalQuillDelta(t *testing.T) {
	var exo, aco []*FormatText
	var quillDelta string
	quillDelta = "{\"ops\":[{\"insert\":\"{user: \"},{\"insert\":{\"dynamicField\":{\"id\":\"jhfjqvfyifrxlotl\",\"value\":\"user\",\"key\":\"user.undefined\"}}},{\"insert\":\"}\"}]}"
	exo = []*FormatText{{Text: "{user: "}, {Key: "user.undefined"}, {Text: "}"}}
	aco = unmarshalQuillDelta(quillDelta)
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("cant unmarshal QuillDelta")
	}

	quillDelta = "[{\"insert\":\"hello \"},{\"insert\":{\"dynamicField\":{\"id\":\"oaqjdmekllkgpclo\",\"value\":\"Họ và tên\",\"key\":\"user.fullname\"}}},{\"insert\":\"  \"},{\"insert\":{\"dynamicField\":{\"id\":\"rqeghvmhqagtapfu\",\"value\":\"Địa chỉ email\",\"key\":\"user.emails\"}}},{\"insert\":\" hi\"}]"
	exo = []*FormatText{{Text: "hello "}, {Key: "user.fullname"}, {Text: "  "}, {Key: "user.emails"}, {Text: " hi"}}
	aco = unmarshalQuillDelta(quillDelta)
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("cant unmarshal QuillDelta")
	}
}

func TestGetText(t *testing.T) {
	var exo, aco string
	var objTree *ObjectTree
	quillDelta := `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`
	var rooti interface{}
	err := json.Unmarshal([]byte(quillDelta), &rooti)
	if err != nil {
		t.Error(err)
		return
	}
	root, ok := rooti.(map[string]interface{})
	if !ok {
		t.Error("cant unmarshal json")
	}
	objTree = &ObjectTree{data: root}
	exo = "user"
	aco = objTree.GetText("ops.1.insert.dynamicField.key")
	if aco != exo {
		t.Errorf("want %s, actual %s", exo, aco)
	}
}

func TestGetKVs(t *testing.T) {
	var exo, aco []*pb.KV
	quillDelta := `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`
	var rooti interface{}
	err := json.Unmarshal([]byte(quillDelta), &rooti)
	if err != nil {
		t.Error(err)
		return
	}
	root, ok := rooti.(map[string]interface{})
	if !ok {
		t.Error("cant unmarshal json")
		return
	}
	objTree := &ObjectTree{data: root}
	exo = []*pb.KV{{Key: "insert", Value: "{\n\t\"user\": "}, {Key: "insert.dynamicField.key", Value: "user"}, {Key: "insert", Value: "\n}\n"}}
	aco = objTree.GetKVs("ops", []string{"insert.dynamicField.key", "insert"})
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %s, actual %s", exo, aco)
	}
}

func TestGjson(t *testing.T) {
	var exo, aco []*pb.KV
	quillDelta := `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`
	insertResults := gjson.Get(quillDelta, "ops").Array()
	aco = make([]*pb.KV, 0)
	for i := 0; i < len(insertResults); i++ {
		if jstr := gjson.Get(insertResults[i].Raw, "insert.dynamicField.key").Str; jstr != "" {
			aco = append(aco, &pb.KV{Key: "insert.dynamicField.key", Value: jstr})
			continue
		}
		if jstr := gjson.Get(insertResults[i].Raw, "insert").Str; jstr != "" {
			aco = append(aco, &pb.KV{Key: "insert", Value: jstr})
			continue
		}
	}
	exo = []*pb.KV{{Key: "insert", Value: "{\n\t\"user\": "}, {Key: "insert.dynamicField.key", Value: "user"}, {Key: "insert", Value: "\n}\n"}}
	if !reflect.DeepEqual(aco, exo) {
		t.Errorf("want %s, actual %s", exo, aco)
	}
}

// win
func BenchmarkTestGjson(b *testing.B) {
	var aco []*pb.KV
	quillDelta := `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`
	for n := 0; n < b.N; n++ {
		insertResults := gjson.Get(quillDelta, "ops").Array()
		aco = make([]*pb.KV, 0)
		for i := 0; i < len(insertResults); i++ {
			if jstr := gjson.Get(insertResults[i].Raw, "insert.dynamicField.key").Str; jstr != "" {
				aco = append(aco, &pb.KV{Key: "insert.dynamicField.key", Value: jstr})
				continue
			}
			if jstr := gjson.Get(insertResults[i].Raw, "insert").Str; jstr != "" {
				aco = append(aco, &pb.KV{Key: "insert", Value: jstr})
				continue
			}
		}
	}
}

func BenchmarkTestGetKVs(b *testing.B) {
	quillDelta := `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`
	var rooti interface{}
	err := json.Unmarshal([]byte(quillDelta), &rooti)
	if err != nil {
		return
	}
	root, ok := rooti.(map[string]interface{})
	if !ok {
		return
	}
	objTree := &ObjectTree{data: root}
	for n := 0; n < b.N; n++ {
		objTree.GetKVs("ops", []string{"insert.dynamicField.key", "insert"})
	}
}

func BenchmarkUnmarshalQuillDelta(b *testing.B) {
	quillDelta := `{"ops":[{"insert":"{\n\t\"user\": "},{"insert":{"dynamicField":{"id":"jhfjqvfyifrxlotl","value":"user","key":"user"}}},{"insert":"\n}\n"}]}`
	for n := 0; n < b.N; n++ {
		unmarshalQuillDelta(quillDelta)
	}
}
