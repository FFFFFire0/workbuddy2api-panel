package upstream

import (
	"encoding/json"
	"testing"
)

// 回迁自 linguo 线的 mergeAdjacentToolCalls 回归测试（DeepSeek 11148 修复）。
// 用独立辅助函数名，避免与 tool_pairing_test.go 的现有 helpers 冲突。

func mjMsgs(t *testing.T, raw string) []any {
	t.Helper()
	var v []any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return v
}

func mjSummary(msgs []any) []string {
	out := make([]string, 0, len(msgs))
	for _, m := range msgs {
		obj, ok := m.(map[string]any)
		if !ok {
			out = append(out, "<non-object>")
			continue
		}
		role, _ := obj["role"].(string)
		if calls, ok := obj["tool_calls"].([]any); ok && len(calls) > 0 {
			ids := ""
			for _, c := range calls {
				cm, _ := c.(map[string]any)
				id, _ := cm["id"].(string)
				ids += id + ","
			}
			out = append(out, role+"("+ids[:len(ids)-1]+")")
		} else if role == "tool" {
			id, _ := obj["tool_call_id"].(string)
			out = append(out, role+"("+id+")")
		} else {
			out = append(out, role+"(-)")
		}
	}
	return out
}

func mjAssert(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("summary = %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("summary = %v want %v", got, want)
		}
	}
}

func TestMJMergeAdjacentParallelToolCalls(t *testing.T) {
	const (
		aC00 = `{"role":"assistant","content":null,"tool_calls":[{"id":"c00","type":"function","function":{"name":"f","arguments":"{}"}}]}`
		aC01 = `{"role":"assistant","content":null,"tool_calls":[{"id":"c01","type":"function","function":{"name":"f","arguments":"{}"}}]}`
		aC02 = `{"role":"assistant","content":null,"tool_calls":[{"id":"c02","type":"function","function":{"name":"f","arguments":"{}"}}]}`
		tC00 = `{"role":"tool","tool_call_id":"c00","content":"r0"}`
		tC01 = `{"role":"tool","tool_call_id":"c01","content":"r1"}`
	)

	t.Run("parallel two merged", func(t *testing.T) {
		in := mjMsgs(t, "["+aC00+","+aC01+","+tC00+","+tC01+"]")
		out, changed := mergeAdjacentToolCalls(in)
		if !changed {
			t.Fatal("changed=false, want merge")
		}
		mjAssert(t, mjSummary(out), []string{"assistant(c00,c01)", "tool(c00)", "tool(c01)"})
		tcs := out[0].(map[string]any)["tool_calls"].([]any)
		if id, _ := tcs[0].(map[string]any)["id"].(string); id != "c00" {
			t.Errorf("first tool_call = %q want c00", id)
		}
	})

	t.Run("three consecutive merged", func(t *testing.T) {
		out, changed := mergeAdjacentToolCalls(mjMsgs(t, "["+aC00+","+aC01+","+aC02+"]"))
		if !changed {
			t.Fatal("changed=false, want merge")
		}
		mjAssert(t, mjSummary(out), []string{"assistant(c00,c01,c02)"})
	})

	t.Run("already normalized no-op", func(t *testing.T) {
		merged := `{"role":"assistant","content":null,"tool_calls":[{"id":"c00"},{"id":"c01"}]}`
		in := mjMsgs(t, "["+merged+","+tC00+","+tC01+"]")
		out, changed := mergeAdjacentToolCalls(in)
		if changed {
			t.Error("changed=true, normalized shape must not change")
		}
		mjAssert(t, mjSummary(out), []string{"assistant(c00,c01)", "tool(c00)", "tool(c01)"})
	})

	t.Run("separated by user message not merged", func(t *testing.T) {
		in := mjMsgs(t, "["+aC00+`,{"role":"user","content":"x"},`+aC01+"]")
		out, changed := mergeAdjacentToolCalls(in)
		if changed {
			t.Error("changed=true, non-adjacent must not merge")
		}
		mjAssert(t, mjSummary(out), []string{"assistant(c00)", "user(-)", "assistant(c01)"})
	})

	t.Run("non-empty content not merged", func(t *testing.T) {
		in := mjMsgs(t, "["+aC00+`,{"role":"assistant","content":"text","tool_calls":[{"id":"c01"}]}]`)
		out, changed := mergeAdjacentToolCalls(in)
		if changed {
			t.Error("changed=true, non-empty content must not merge")
		}
		mjAssert(t, mjSummary(out), []string{"assistant(c00)", "assistant(c01)"})
	})

	t.Run("reasoning_content appended", func(t *testing.T) {
		in := mjMsgs(t, `[{"role":"assistant","content":null,"reasoning_content":"r1","tool_calls":[{"id":"c00"}]},{"role":"assistant","content":null,"reasoning_content":"r2","tool_calls":[{"id":"c01"}]}]`)
		out, changed := mergeAdjacentToolCalls(in)
		if !changed {
			t.Fatal("changed=false, want merge")
		}
		rc, _ := out[0].(map[string]any)["reasoning_content"].(string)
		if rc != "r1\nr2" {
			t.Errorf("reasoning_content = %q want %q", rc, "r1\nr2")
		}
	})

	t.Run("reverse shape text folded in", func(t *testing.T) {
		in := mjMsgs(t, `[{"role":"assistant","content":null,"tool_calls":[{"id":"c00"}]},{"role":"assistant","content":"done"}]`)
		out, changed := mergeAdjacentToolCalls(in)
		if !changed {
			t.Fatal("changed=false, want fold")
		}
		obj := out[0].(map[string]any)
		if obj["content"] != "done" {
			t.Errorf("content = %v want done", obj["content"])
		}
	})

	t.Run("empty array content counts as empty", func(t *testing.T) {
		in := mjMsgs(t, `[{"role":"assistant","content":[],"tool_calls":[{"id":"c00"}]},`+aC01+"]")
		_, changed := mergeAdjacentToolCalls(in)
		if !changed {
			t.Error("changed=false, empty array content must merge")
		}
	})
}
