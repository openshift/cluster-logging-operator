package luaunittest

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

const (
	//Dropped a record is dropped
	Dropped = -1
	//Unmodified means no changes to the record
	Unmodified = 0
	//ModifiedTimestampAndRecord the result has modified timestamp and record
	ModifiedTimestampAndRecord = 1
	//ModifiedRecord the result only modifies the record
	ModifiedRecord = 2
)

type Record map[string]interface{}

func (rec Record) SetParialRecord() {
	rec["logtag"] = "P"
}
func (rec Record) SetFullRecord() {
	rec["logtag"] = "F"
}
func (rec Record) Message() string {
	return rec["message"].(string)
}

//String is written to support serializing a record to be compatible with a lua script
func (rec Record) String() string {
	entries := []string{}
	for k, v := range rec {
		var template string
		valueType := reflect.TypeOf(v)
		switch valueType.Kind() {
		case reflect.Int:
			template = "%s=%d"
		default:
			template = "%s=%q"
		}
		entries = append(entries, fmt.Sprintf(template, k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(entries, ","))
}

//Runner executes a script using a given set of inputs and collects the resuls
type Runner struct {
	script     string
	filterName string
	Inputs     []Input
	Results    []Result
}

//Input is set of parameters given to a filter
type Input struct {
	Tag       string
	Timestamp float64
	Record    Record
}

//String is written to support serializing a record to be compatible with a lua script
func (in Input) String() string {
	return fmt.Sprintf("{tag=%q,timestamp=%f,record=%s}", in.Tag, in.Timestamp, in.Record)
}

//Result are the items returned by a filter
type Result struct {
	Code      int
	Timestamp float64
	Record    Record
}

func newResult(luaValue *lua.LTable) Result {
	option := gluamapper.Option{NameFunc: func(name string) string { return name }}
	result := Result{}
	luaValue.ForEach(func(key, value lua.LValue) {
		if key.Type() == lua.LTString {
			if value == nil {
				return
			}
			v := gluamapper.ToGoValue(value, option)
			entryKey := gluamapper.ToGoValue(key, option).(string)
			switch entryKey {
			case "code":
				result.Code = int(v.(float64))
			case "timestamp":
				result.Timestamp = v.(float64)
			case "record":
				if reflect.TypeOf(v).Kind() == reflect.Map {
					resultRecord := map[string]interface{}{}
					record := v.(map[interface{}]interface{})
					for k, recordValue := range record {
						resultRecord[k.(string)] = recordValue
					}
					result.Record = Record(resultRecord)
				}
			}
		}
	})
	return result
}

var runnerScriptTemplate string = `
inputs = {%s}
results = {}
for i=1,%d do
	input = inputs[i]
	code, timestamp, record = %s(input.tag, input.timestamp, input.record)
	result = {}
	result['code'] = code
	result['timestamp'] = timestamp
	result['record'] = record
	table.insert(results, result)
end
`

func NewRunner(script, filterName string, inputs ...Input) (*Runner, error) {
	if !strings.Contains(script, fmt.Sprintf("function %s", filterName)) {
		return nil, fmt.Errorf("The script does not include a funcation named %s:", filterName)
	}
	return &Runner{
		script:     script,
		filterName: filterName,
		Inputs:     inputs,
		Results:    []Result{},
	}, nil
}

func (run *Runner) Run() error {
	vm := lua.NewState()
	defer vm.Close()
	inputs := []string{}
	for _, in := range run.Inputs {
		inputs = append(inputs, in.String())
	}
	runnerScript := fmt.Sprintf(runnerScriptTemplate,
		strings.Join(inputs, ","),
		len(run.Inputs),
		run.filterName,
	)

	script := fmt.Sprintf("%s\n%s", run.script, runnerScript)
	if err := vm.DoString(script); err != nil {
		fmt.Println(runnerScript)
		return err
	}
	results := vm.GetGlobal("results")
	resultTable := results.(*lua.LTable)
	resultTable.ForEach(func(key, value lua.LValue) {
		entry := value.(*lua.LTable)
		run.Results = append(run.Results, newResult(entry))
	})
	return nil
}
