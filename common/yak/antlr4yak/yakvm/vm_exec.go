package yakvm

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/mutate"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm/vmstack"
	"github.com/yaklang/yaklang/common/yakdocument"
	"reflect"
	"strconv"
	"strings"
)

type ExitCodeType int

const (
	ErrorExit ExitCodeType = 1 << iota
	NormallyExit
	NoneExit
	TryExit
	ReturnExit
)

func (v *Frame) nextCode() {
	v.codePointer++
}

func (v *Frame) setCodeIndex(i int) {
	if v.vm.debugMode && v.vm.debugger.jmpIndex == -1 {
		v.vm.debugger.jmpIndex = i
	}

	v.codePointer = i
}

func (v *Frame) Exec(codes []*Code) {
	v.codes = codes
	v.execEx()
}
func (v *Frame) DebugExec(codes []*Code) {
	v.debug = true
	v.codes = codes
	v.execEx()
}

func (v *Frame) NormalExec(codes []*Code) {
	v.debug = false
	v.codes = codes
	v.execEx()
}
func (v *Frame) continueExec() {
	v.execExWithContinueOption(true)
}
func (v *Frame) execEx() {
	v.execExWithContinueOption(false)
}
func (v *Frame) execExWithContinueOption(isContinue bool) {
	if !isContinue {
		v.codePointer = 0
	}
	//退出代码 // -1代表异常，0代表代码执行到最后，1代表通过panic、return等方式退出
	v.exitCode = ErrorExit
	deferStack := vmstack.New()

	defer func() {
		if v.codePointer >= len(v.codes) {
			v.exitCode = NormallyExit
		}

		if v.debug && v.exitCode == NormallyExit {
			if v.stack.Len() > 0 {
				log.Warnf("Runtime Stack Unbalanced: %v", v.stack.Len())
				for v.stack.Len() > 0 {
					value := v.stack.Pop().(*Value).Value
					switch vv := value.(type) {
					case []*Value:
						for _, val := range vv {
							fmt.Printf("stack - : %#v", val)
						}
					default:
						fmt.Printf("stack : %#v\n", value)
					}
				}

				panic("Runtime Stack Unbalanced!")
			}

			if v.iteratorStack.Len() > 0 {
				log.Warnf("Iterator Stack Unbalanced: %v", v.iteratorStack.Len())
				panic("Runtime For-Iterator-Stack Unbalanced!")
			}

			if !v.CurrentScope().IsRoot() {
				log.Warnf("Scope is unbalanced")
			}
		}

		//defer 中捕获 panic 并抛出
		var (
			ierr    interface{}
			vmPanic *VMPanic
			ok      bool
		)

		// indebuggerEval就不处理panic了，直接抛出
		if !v.vm.GetConfig().GetStopRecover() {
			if ierr = recover(); ierr != nil && !v.indebuggerEval {
				if vmPanic, ok = ierr.(*VMPanic); !ok {
					// go产生的panic需要套一层VMPanic
					vmPanic = NewVMPanic(ierr)
				}
				v.panic(vmPanic)
			}
		}
		if !v.vm.GetConfig().GetStopRecover() {
			if i := recover(); i != nil {
				switch ret := i.(type) {
				case *VMPanic: // 由yaklang产生的panic
					v.panic(ret)
				default: // go产生的panic
					v.panic(NewVMPanic(i))
				}
				v.panic(vmPanic)
			}

		}
		scopeBack := v.scope
		returnVal := v.lastStackValue
		//执行defer中的代码
		for {
			if deferStack.Len() <= 0 {
				break
			}
			deferCodes := deferStack.Pop().(*Defer)
			v.scope = deferCodes.Scope
			v.Exec(deferCodes.Codes)
		}
		v.scope = scopeBack
		v.lastStackValue = returnVal

		if v.vm.debugMode {
			if v.indebuggerEval {
				if ierr != nil {
					// 在debugger中执行代码,直接panic由debugger捕获
					panic(ierr)
				}
				return
			}
			if vmPanic != nil {
				log.Error(vmPanic)
			}
			debugger := v.vm.debugger
			if debugger.InRootState() {
				debugger.description = "The program is finished"
				debugger.finished = true

				v.vm.debugger.Callback()
			}
		}
	}()

	if v.debug {
		ShowOpcodes(v.codes)
		println()
		println(strings.Repeat("-", 32))
		println()
	}
	for {
		if v.codePointer >= len(v.codes) {
			v.exitCode = NormallyExit
			return
		}
		var code = v.codes[v.codePointer]
		if v.debug {
			fmt.Printf("%-16s %4d:%v\n", code.RangeVerbose(), v.codePointer, code.String())
		}
		//for _, fun := range v.vm.BreakPoint {
		//	fun(v)
		//}
		if code.Opcode == OpStopCatchError {
			v.exitCode = TryExit
			return
		}
		v.execCode(code, v.debug)

		opCodeFlag := code.Opcode
		if opCodeFlag == OpDefer {
			//  添加到 deferStack 中
			deferStack.Push(&Defer{
				Codes: code.Op1.Value.([]*Code),
				Scope: v.CurrentScope(),
			})
		} else if opCodeFlag.IsJmp() {
			continue
		} else if opCodeFlag == OpReturn || opCodeFlag == OpPanic {
			v.exitCode = ErrorExit
			return
		}
		v.nextCode()
	}
}

func ShowOpcodes(c []*Code) {
	for i, code := range c {
		fmt.Printf("%-13s %4d:%v\n", code.RangeVerbose(), i, code.String())
	}
}

func OpcodesString(c []*Code) string {
	var buf strings.Builder
	for i, code := range c {
		buf.WriteString(fmt.Sprintf("%-13s %4d:%v\n", code.RangeVerbose(), i, code.String()))
	}
	return buf.String()
}

var undefined = &Value{
	TypeVerbose: "undefined",
	Value:       nil,
	Literal:     "undefined",
}

func IsUndefined(v *Value) bool {
	return v == undefined
}

func GetUndefined() *Value {
	return undefined
}

func NewUndefined(id int) *Value {
	return &Value{
		TypeVerbose: "undefined",
		Value:       nil,
		Literal:     "undefined",
		SymbolId:    id,
	}
}

func (v *Frame) execCode(c *Code, debug bool) {
	select {
	case <-v.ctx.Done():
		//log.Warn("YakVM Frame Exec Code Terminated by Context Control")
		v.codePointer = len(v.codes)
		return
	default:
		v._execCode(c, debug)
	}
}
func (v *Frame) _execCode(c *Code, debug bool) {
	//if v.codePointer == 40 {
	//	println()
	//}
	if v.vm.debugMode && !v.indebuggerEval {
		debugger := v.vm.debugger
		debugger.Wait()
		debugger.BreakPointCallback(v.codePointer)
	}
	// defer func() {
	// 	// 如果debugger想要步过且code确实jmp了,则在其jmp位置设置临时断点
	// 	// if v.vm.debugMode && v.vm.debugger.nextState != nil && v.vm.debugger.jmpIndex >= 0 {
	// 	// 	v.vm.debugger.HandleForStepNextJmp(c)
	// 	// }
	// }()

	switch c.Opcode {
	case OpCatchError:
		catchCodeIndex := c.Op1.Int()
		id := c.Op2.Int()
		v.catchErrorRun(catchCodeIndex, id)
	case OpScope:
		v.CreateAndSwitchSubScope(v.scope.symtbl.GetSymbolTableById(c.Unary))
	case OpScopeEnd:
		v.ExitScope()
	case OpAsyncCall:
		// 底层调用 goroutine 来进行操作的
		wavy := false
		if c.Op1 != nil {
			wavy = c.Op1.Bool()
		}
		args := v.popArgN(c.Unary)
		callableValue := v.pop()
		v.asyncCall(callableValue, wavy, args)
	case OpAssign:
		switch v.vm.GetConfig().vmMode {
		case NASL:
			if c.Op1 != nil && c.Op1.IsString() && c.Op1.String() == "nasl_global_declare" {
				valId := v.pop()
				_, ok := v.CurrentScope().GetValueByID(valId.SymbolId)
				if !ok {
					valId.AssignBySymbol(v.CurrentScope(), GetUndefined())
				}
				return
			}
			if c.Op1 != nil && c.Op1.IsString() && c.Op1.String() == "nasl_declare" {
				valId := v.pop()
				_, ok := v.CurrentScope().GetValueByID(valId.SymbolId)
				if !ok {
					valId.AssignBySymbol(v.CurrentScope(), GetUndefined())
				}
				return
			}
			arg2 := v.pop()
			arg1 := v.pop()
			if arg1.IsLeftValueRef() {
				arg1.AssignBySymbol(v.CurrentScope(), arg2)
			} else {
				if v1, ok := arg1.Value.([]*Value); ok {
					if len(v1) == 2 {
						if v1[0].IsUndefined() {
							if v1[0].SymbolId > 0 {
								initMap := NewAutoValue(map[interface{}]interface{}{})
								v.CurrentScope().NewValueByID(v1[0].SymbolId, initMap)
								v1[0] = initMap
							} else {
								panic("unreasonable undefined")
							}
						}
					}
				}
				arg1.Assign(v, arg2)
			}
			v.push(arg2)
			return
		case LUA:
			assignArgs := v.popArgN(2)
			leftValues := assignArgs[1]
			rightValues := assignArgs[0]
			rightVal := rightValues.Value
			leftVal := leftValues.Value

			if v, ok := rightVal.([]*Value); ok && len(v) > 0 {
				rightVal = v[0].Value
			}
			if _, ok := rightVal.(*Function); ok {
				if rightVal.(*Function).scope == nil {
					rightVal.(*Function).scope = v.CurrentScope()
				}
				if val, ok := leftVal.([]*Value); ok && len(val) > 0 {
					lv, err := val[0].ConvertToLeftValue()
					if err == nil {
						if lv.IsLeftValueRef() {
							funcName, ok := v.CurrentScope().symtbl.GetNameByVariableId(lv.SymbolId)
							if ok {
								rightVal.(*Function).anonymousFunctionBindName = funcName
							}
						}
					}
				}
			}
			if c.Unary == 0 {
				v.luaGlobalAssign(leftValues, rightValues)
			} else {
				v.luaLocalAssign(leftValues, rightValues)
			}
			return
		case YAK:
			fallthrough
		default:
			assignArgs := v.popArgN(2)
			leftValues := assignArgs[1]
			rightValues := assignArgs[0]
			rightVal := rightValues.Value
			leftVal := leftValues.Value

			if v, ok := rightVal.([]*Value); ok && len(v) > 0 {
				rightVal = v[0].Value
			}
			if _, ok := rightVal.(*Function); ok {
				rightVal.(*Function).scope = v.CurrentScope()
				if val, ok := leftVal.([]*Value); ok && len(val) > 0 {
					lv, err := val[0].ConvertToLeftValue()
					if err == nil {
						if lv.IsLeftValueRef() {
							funcName, ok := v.CurrentScope().symtbl.GetNameByVariableId(lv.SymbolId)
							if ok {
								rightVal.(*Function).anonymousFunctionBindName = funcName
							}
						}
					}
				}
			}
			//if leftVal
			v.assign(leftValues, rightValues)
			return
		}
	case OpFastAssign:
		switch v.vm.GetConfig().vmMode {
		case LUA:
			args := v.popArgN(2)
			arg1, arg2 := args[0], args[1]
			if c.Unary == 0 {
				arg1.GlobalAssignBySymbol(v.CurrentScope(), arg2)
			} else {
				arg1.AssignBySymbol(v.CurrentScope(), arg2)
			}
			v.push(arg2)
			return
		case YAK:
			fallthrough
		default:
			args := v.popArgN(2)
			arg1, arg2 := args[0], args[1]
			arg1.AssignBySymbol(v.CurrentScope(), arg2)
			v.push(arg2)
			return
		}
	case OpList:
		if c.Unary <= 0 {
			return
		}
		v.push(&Value{
			TypeVerbose: "__opcode_list__",
			Value:       v.popArgN(c.Unary),
		})
	case OpRangeNext:
		var jmp bool

		iterator := v.iteratorStack.Peek().(IteratorInterface)
		if iterator == nil {
			panic("iterator is nil")
		}

		if c.Unary <= 0 {
			return
		}

		var values = make([]*Value, c.Unary)
		for i := 0; i < c.Unary; i++ {
			values[i] = GetUndefined()
		}
		data, hadEnd := iterator.Next()
		typ := iterator.Type()

		if c.Unary > 1 && typ == ChannelIteratorType {
			panic(fmt.Sprintf("range over channel allow only one iteration variable"))
		}

		min := c.Unary
		if len(data) < min {
			min = len(data)
		}

		for i := 0; i < min; i++ {
			values[i] = NewAutoValue(data[i])
		}

		if typ == ChannelIteratorType && hadEnd {
			jmp = true
			v.setCodeIndex(c.Op1.Int())
		}

		if !jmp {
			v.push(&Value{
				TypeVerbose: "__opcode_list__",
				Value:       values,
			})
			v.nextCode()
		}
		return
	case OpInNext:
		var jmp bool

		iterator := v.iteratorStack.Peek().(IteratorInterface)
		if iterator == nil {
			panic("iterator is nil")
		}

		if c.Unary <= 0 {
			return
		}

		var values = make([]*Value, c.Unary)
		for i := 0; i < c.Unary; i++ {
			values[i] = GetUndefined()
		}
		data, hadEnd := iterator.Next()
		typ := iterator.Type()

		if c.Unary > 1 && typ == ChannelIteratorType {
			panic(fmt.Sprintf("range over channel allow only one iteration variable"))
		}

		// 对sliceIterator做处理，第一个索引应该抛弃
		if typ == SliceIteratorType && len(data) > 1 {
			data = data[1:]
		}
		// 对sliceIterator做解包处理
		value := data[0]
		firstReflectValue := reflect.ValueOf(value)
		firstReflectValueKind := firstReflectValue.Kind()
		// 只有当左值数量大于1的时候才解包
		if c.Unary > 1 && (firstReflectValueKind == reflect.Slice || firstReflectValueKind == reflect.Array) {
			reflectValueLen := firstReflectValue.Len()
			if reflectValueLen != c.Unary {
				panic(fmt.Sprintf("values(%v) length not equal to unpacking length, got %d, want %d", value, reflectValueLen, c.Unary))
			}

			// 解包
			newData := make([]interface{}, reflectValueLen)
			for i := 0; i < reflectValueLen; i++ {
				newData[i] = firstReflectValue.Index(i).Interface()
			}
			data = newData
		}

		min := c.Unary
		if len(data) < min {
			min = len(data)
		}

		for i := 0; i < min; i++ {
			values[i] = NewAutoValue(data[i])
		}

		if hadEnd && typ == ChannelIteratorType {
			jmp = true
			v.setCodeIndex(c.Op1.Int())
		}

		if !jmp {
			v.push(&Value{
				TypeVerbose: "__opcode_list__",
				Value:       values,
			})
			v.nextCode()
		}

		return
	case OpEnterFR:
		op := v.peek()
		iterator, err := NewIterator(op.Value)
		if err != nil {
			panic(fmt.Sprintf("%#v is not rangeable", op.TypeVerbose))
		} else if rv := reflect.ValueOf(iterator); !rv.IsValid() || rv.IsNil() {
			v.setCodeIndex(c.Unary)
			return
		} else {
			v.nextCode()
		}
		v.iteratorStack.Push(iterator)
		return
	case OpExitFR:
		iterator := v.iteratorStack.Peek().(IteratorInterface)

		if !iterator.IsEnd() {
			v.setCodeIndex(c.Unary)
		} else {
			v.iteratorStack.Pop()
			v.nextCode()
		}
		return
	case OpPush:
		v.GetLastStackValue()
		if c.Op1.IsStringOrBytes() {
			strValue := c.Op1.AsString()
			switch c.Unary {
			case 0:
				v.push(NewStringValue(strValue))
			case 'b':
				v.push(NewAutoValue([]byte(strValue)))
			case 'x':
				//使用了f前缀生成的是 string slice
				value, err := mutate.FuzzTagExec(strValue)
				if err != nil {
					v.push(NewStringSliceValue([]string{}))
					log.Error(err)
					//解析fuzztag出错时不panic，防止fuzztag解析失败导致语言引擎异常
					//panic(fmt.Sprintf("mutate.FuzzTagExec failed: %s", err))
				}
				v.push(NewStringSliceValue(value))
			default:
				panic("unknown string prefix")
			}
			return
		}
		if c.Op1.IsYakFunction() {
			fun := c.Op1.Value.(*Function)
			fun.scope = v.scope
		}
		v.push(c.Op1)
		return
	case OpPushLeftRef:
		// 通过符号寻找对应的值
		id := c.Unary
		v.push(NewValueRef(id))
		return
	case OpPushRef:
		switch v.vm.GetConfig().vmMode {
		case NASL:
			id := c.Unary
			val, ok := v.CurrentScope().GetValueByID(id)
			if !ok {
				name, ok1 := v.CurrentScope().GetSymTable().GetNameByVariableId(id)
				if ok1 {
					if v, ok1 := v.GlobalVariables[name]; ok1 {
						val = NewValue("function", v, name)
						ok = true
					} else {
						panic("BUG: cannot found value by name:[" + name + "]")
					}
				}
			}
			if !ok {
				panic("BUG: cannot found value by symbol:[" + fmt.Sprint(id) + "]")
			}
			if val.Value == nil {
				val = NewUndefined(id)
			}
			v.push(val)
			return
		case YAK:
			id := c.Unary
			val, ok := v.CurrentScope().GetValueByID(id)
			if !ok {
				panic(utils.Errorf("BUG: cannot found value by symbol [%d: %s]", id, v.CurrentScope().GetNameById(id)))
			}
			//t, err := v.RootSymbolTable.FindSymbolTableBySymbolId(id)
			//if err != nil {
			//	panic(err)
			//}
			//val, _ := t.GetValueByVariableId(id)
			if val.Value == nil {
				v.push(undefined)
			} else {
				v.push(val)
			}
			return
		}
	case OpPushId:
		// 不是赋值的符号，一般来说，这应该是外部注入的变量或者函数
		// 通过 GlobalVariables 可以很棒的处理这个问题
		name := c.Op1.String()
		value, ok := v.GlobalVariables[name]
		if !ok {
			id, ok := v.CurrentScope().symtbl.GetSymbolByVariableName(name)
			if ok {
				val, ok := v.CurrentScope().GetValueByID(id)
				if ok {
					v.push(val)
					return
				}
			}
			log.Warnf("yakvm: cannot found variable: %v, use undefined", c.Op1.String())
			v.push(undefined)
			return
		} else {
			val := &Value{
				TypeVerbose: "global",
				Literal:     name,
				Value:       value,
			}
			val.AddExtraInfo("global", true)
			v.push(val)
		}
	case OpTypeCast:
		typ := v.pop()
		val := v.pop()
		var resVal *Value
		var err error

		//if typeN == 2 { //进行复合类型判断
		//	//转换不成功继续尝试 slice转换
		//	switch types[0].Type().Kind() {
		//	case reflect.Slice:
		//		switch types[1].Type() {
		//		case literalReflectType_Byte:
		//			if val.IsString() {
		//				resVal = NewValue("byte", []byte(val.String()), val.String())
		//			}
		//
		//		}
		//	}
		//	if resVal == nil {
		//		panic(fmt.Sprintf("cannot cast %v to []%v", val.Value, types[1].Type().String()))
		//	}
		//	v.push(resVal)
		//} else { //基本类型判断

		// undefined 处理，返回零值
		if val.IsUndefined() {
			rv := reflect.New(typ.Type())
			v.push(NewAutoValue(rv.Elem().Interface()))
			return
		}

		//基本类型互转
		switch typ.Type() {
		case literalReflectType_Int:
			if val.IsInt() {
				resVal = val
			}
			if val.IsFloat() {
				resVal = NewIntValue(int(val.Float64()))
			}
			if val.IsBool() {
				if val.Bool() {
					resVal = NewIntValue(1)
				} else {
					resVal = NewIntValue(0)
				}
			}
			if val.IsStringOrBytes() {
				intVal, e := strconv.Atoi(val.String())
				if e != nil {
					err = e
					break
				}
				resVal = NewIntValue(intVal)
			}
			if val.IsByte() {
				resVal = NewIntValue(int(val.Value.(byte)))
			}
		case literalReflectType_Bool:
			if val.IsBool() {
				resVal = val
			}
			if val.IsStringOrBytes() {
				if val.String() == "false" {
					resVal = NewBoolValue(false)
				} else {
					resVal = NewBoolValue(true)
				}
			}
			if val.IsInt() {
				resVal = NewBoolValue(val.Int64() != 0)
			}
			if val.IsFloat() {
				resVal = NewBoolValue(val.Float64() != 0)
			}
			if val.IsByte() {
				resVal = NewBoolValue(val.Value.(byte) != 0)
			}
		case literalReflectType_Float64:
			if val.IsFloat() {
				resVal = val
			}
			if val.IsInt() {
				resVal = NewValue("float64", float64(val.Int64()), fmt.Sprint(val.Int64()))
			}
			if val.IsBool() {
				if val.Bool() {
					resVal = NewValue("float64", float64(1), "1")
				} else {
					resVal = NewValue("float64", float64(0), "0")
				}
			}
			if val.IsStringOrBytes() {
				floatVal, e := strconv.ParseFloat(val.String(), 64)
				if e != nil {
					err = e
					break
				}
				resVal = NewValue("float64", floatVal, val.String())
			}
			if val.IsByte() {
				resVal = NewValue("float64", float64(val.Value.(byte)), fmt.Sprint(val.Value.(byte)))
			}
		case literalReflectType_Byte:
			if val.IsByte() {
				resVal = val
			}
			if val.IsInt() {
				resVal = NewValue("byte", byte(val.Int64()), fmt.Sprint(val.Int64()))
			}
			if val.IsBool() {
				if val.Bool() {
					resVal = NewValue("byte", byte(1), "1")
				} else {
					resVal = NewValue("byte", byte(0), "0")
				}
			}
			if val.IsStringOrBytes() {
				byteVal, e := strconv.Atoi(val.String())
				if e != nil {
					err = e
					break
				}
				resVal = NewValue("byte", byte(byteVal), val.String())
			}
			if val.IsFloat() {
				resVal = NewValue("byte", byte(val.Float64()), fmt.Sprint(val.Float64()))
			}
		case literalReflectType_String:
			resVal = NewStringValue(val.String())
		case literalReflectType_Interface:
			resVal = NewAutoValue(val.Value)
		}
		if err != nil {
			panic(fmt.Sprintf("cannot cast %v to %v,err: %v", val.Value, typ.Type().String(), err))
		}
		if resVal != nil {
			v.push(resVal)
			return
		}
		//如果未成功解析基本类型，则继续尝试解析复合类型
		switch typ.Type().String() {
		case "[]uint8":
			if val.IsStringOrBytes() {
				resVal = NewValue("[]type", []byte(val.String()), "")
			}
		}
		if resVal != nil {
			v.push(resVal)
			return
		}
		//还不成功就只能试试自动转换了，转换失败就panic
		refV := reflect.ValueOf(val)
		err = v.AutoConvertReflectValueByType(&refV, typ.Type())
		if err != nil {
			panic(fmt.Sprintf("cannot cast %v to %v", val.Value, typ.Type().String()))
		} else {
			resVal = NewValue("", refV.Interface(), fmt.Sprint(refV.Interface()))
			v.push(resVal)
		}
	case OpNot, OpNeg, OpPlus, OpBitwiseNot:
		op := v.pop()
		op = ChannelValueListToValue(op)
		v.push(v.execOp1(c.Opcode, op))
	case OpPlusPlus, OpMinusMinus:
		var (
			value *Value
			ok    bool
			err   error
		)
		op := v.pop()
		op = ChannelValueListToValue(op)
		op, err = op.ConvertToLeftValue()
		if err != nil {
			panic(fmt.Sprintf("cannot assign value[%v]: %v", op, err))
		}

		scope := v.CurrentScope()
		if op.IsValueList() { // 处理左值是iterablecall的情况
			value = v.getValueForLeftIterableCall(op.ValueList())
		} else if value, ok = scope.GetValueByID(op.SymbolId); !ok && value != undefined {
			if name, ok := scope.GetValueByID(op.SymbolId); ok {
				panic(fmt.Sprintf("cannot get variable[%v] value", name))
			} else {
				panic(fmt.Sprintf("cannot get variable-id[%v] value", op.SymbolId))
			}
		}

		var ret *Value
		switch c.Opcode {
		case OpPlusPlus:
			ret = v.execOp2(OpAdd, value, NewInt64Value(1))
		case OpMinusMinus:
			ret = v.execOp2(OpSub, value, NewInt64Value(1))
		}
		// 将结果重新赋值
		op.Assign(v, ret)
	case OpChan:
		op := v.pop()
		if op.IsChannel() {
			rv := reflect.ValueOf(op.Value)
			if val, ok := rv.Recv(); ok {
				v.push(&Value{
					TypeVerbose: "__channel__opcode_list__",
					Value: []interface{}{
						val.Interface(),
						ok,
					},
				})
				return
			} else {
				v.push(&Value{
					TypeVerbose: "__channel__opcode_list__",
					Value: []interface{}{
						nil,
						ok,
					},
				})
			}
			return
		} else {
			panic(fmt.Sprintf("cannot support chan op1[%v]", op.TypeVerbose))
		}
	case OpAnd, OpAndNot, OpOr, OpXor, OpShl, OpShr, OpAdd, OpSub, OpMul, OpDiv, OpMod,
		OpLtEq, OpLt, OpGtEq, OpGt, OpNotEq, OpEq, OpSendChan, OpIn:
		/* 一般的二元表达式：语言本身没有任何计算能力 */
		args := v.popArgN(2)
		op1, op2 := args[0], args[1]
		op1, op2 = ChannelValueListToValue(op1), ChannelValueListToValue(op2)
		if len(args) != 2 {
			panic("binary op error")
		}

		v.push(v.execOp2(c.Opcode, op1, op2))
		return
	case OpPlusEq, OpMinusEq, OpMulEq, OpDivEq, OpModEq, OpAndEq, OpAndNotEq, OpOrEq, OpXorEq, OpShlEq, OpShrEq:
		var (
			value *Value
			ok    bool
			err   error
		)

		args := v.popArgN(2)
		op1, op2 := args[1], args[0]
		op1, op2 = ChannelValueListToValue(op1), ChannelValueListToValue(op2)
		op1, err = op1.ConvertToLeftValue()
		if err != nil {
			panic(fmt.Sprintf("cannot assign value[%v]: %v", op1, err))
		}

		scope := v.CurrentScope()
		if op1.IsValueList() { // 处理左值是iterablecall的情况
			value = v.getValueForLeftIterableCall(op1.ValueList())
		} else if value, ok = scope.GetValueByID(op1.SymbolId); !ok && value != undefined {
			if name, ok := scope.GetValueByID(op1.SymbolId); ok {
				panic(fmt.Sprintf("cannot get variable[%v] value", name))
			} else {
				panic(fmt.Sprintf("cannot get variable-id[%v] value", op1.SymbolId))
			}
		}

		var ret *Value
		switch c.Opcode {
		case OpPlusEq:
			ret = v.execOp2(OpAdd, value, op2)
		case OpMinusEq:
			ret = v.execOp2(OpSub, value, op2)
		case OpMulEq:
			ret = v.execOp2(OpMul, value, op2)
		case OpDivEq:
			ret = v.execOp2(OpDiv, value, op2)
		case OpModEq:
			ret = v.execOp2(OpMod, value, op2)
		case OpAndEq:
			ret = v.execOp2(OpAnd, value, op2)
		case OpAndNotEq:
			ret = v.execOp2(OpAndNot, value, op2)
		case OpOrEq:
			ret = v.execOp2(OpOr, value, op2)
		case OpXorEq:
			ret = v.execOp2(OpXor, value, op2)
		case OpShlEq:
			ret = v.execOp2(OpShl, value, op2)
		case OpShrEq:
			ret = v.execOp2(OpShr, value, op2)
		default:
			panic(fmt.Sprintf("invalid opcode[%s]", c.String()))
		}
		// 将结果重新赋值
		op1.Assign(v, ret)
	case OpJMP: /*note: control index by your self!*/
		v.setCodeIndex(c.Unary)
		return
	case OpContinue:
		if c.Unary <= 0 {
			panic("runtime error: `continue` should be initialized in `for`")
		}
		v.setCodeIndex(c.Unary)
		v.ExitScopeWithCount(c.Op1.Int())
		return
	case OpBreak:
		if c.Unary <= 0 {
			panic("runtime error: `break` should be initialized in `switch` or `for`")
		}
		v.setCodeIndex(c.Unary)
		switch c.Op2.Int() {
		case 1:
			v.iteratorStack.Pop()
			v.pop()
		}
		v.ExitScopeWithCount(c.Op1.Int())
		// v.ExitScope()
		return
	case OpJMPT: /*note: control index by your self!*/
		switch v.vm.GetConfig().vmMode {
		case LUA:
			if v.pop().LuaTrue() {
				v.setCodeIndex(c.Unary)
			} else {
				v.nextCode()
			}
			return
		case YAK:
			fallthrough
		default:
			if v.pop().True() {
				v.setCodeIndex(c.Unary)
			} else {
				v.nextCode()
			}
			return
		}
	case OpJMPF: /*note: control index by your self!*/
		switch v.vm.GetConfig().vmMode {
		case LUA:
			if v.pop().LuaFalse() {
				v.setCodeIndex(c.Unary)
			} else {
				v.nextCode()
			}
			return
		case YAK:
			fallthrough
		default:
			if v.pop().False() {
				v.setCodeIndex(c.Unary)
			} else {
				v.nextCode()
			}
			return
		}
	case OpJMPTOP: /*note: control index by your self!*/
		switch v.vm.GetConfig().vmMode {
		case LUA:
			op := v.peek()
			if op.LuaTrue() {
				v.setCodeIndex(c.Unary)
			} else {
				v.pop()
				v.nextCode()
			}
			return
		case YAK:
			fallthrough
		default:
			op := v.peek()
			if op.True() {
				v.setCodeIndex(c.Unary)
			} else {
				v.pop()
				v.nextCode()
			}
			return
		}
	case OpJMPFOP: /*note: control index by your self!*/
		switch v.vm.GetConfig().vmMode {
		case LUA:
			op := v.peek()
			if op.LuaFalse() {
				v.setCodeIndex(c.Unary)
			} else {
				v.pop()
				v.nextCode()
			}
			return
		case YAK:
			fallthrough
		default:
			op := v.peek()
			if op.False() {
				v.setCodeIndex(c.Unary)
			} else {
				v.pop()
				v.nextCode()
			}
			return
		}
	case OpCall:
		switch v.vm.GetConfig().vmMode {
		case NASL:
			args := make([]*Value, c.Unary)
			i := c.Unary - 1
			for i >= 0 {
				arg := v.pop()
				if arg.TypeVerbose == "ref" {
					args[i+1].SymbolId = arg.SymbolId
				} else {
					args[i] = arg
					i--
				}
			}
			Val := v.pop()
			var idValue *Value
			if Val.TypeVerbose == "ref" {
				args[i+1].SymbolId = Val.SymbolId
				idValue = v.pop()
			} else {
				idValue = Val
			}
			//内置函数直接调用
			if idValue.Literal == "__method_proxy__" {
				params := [][]interface{}{}
				params = append(params, []interface{}{args[0].Value})
				for i := 1; i < len(args); i++ {
					params = append(params, []interface{}{args[i].SymbolId, args[i].Value})
				}
				val := idValue.Call(v, false, NewValue("", params, ""))
				typeVerbose := "undefined"
				if val != nil {
					typeVerbose = reflect.TypeOf(val).String()
				}
				v.push(&Value{
					TypeVerbose: typeVerbose,
					Value:       val,
				})
			} else {
				if v1, ok := v.GlobalVariables["__OpCallCallBack__"]; ok {
					if v2, ok := v1.(func(string)); ok {
						if v3, ok := idValue.Value.(*Function); ok {
							v2(v3.name)
						}
					}
				}
				val := idValue.Call(v, false, args...)
				typeVerbose := "undefined"
				if val != nil {
					typeVerbose = reflect.TypeOf(val).String()
				}
				v.push(&Value{
					TypeVerbose: typeVerbose,
					Value:       val,
				})
			}
			return
			//外置函数手动调用
			//symbolTable := v.CurrentScope().GetSymTable()
			//funName, ok := symbolTable.GetNameByVariableId(idValue.Int())

			return
		case LUA:
			args := v.popArgN(c.Unary)
			callableValue := v.pop()
			v.callLua(callableValue, args)
		case YAK:
			fallthrough
		default:
			// 函数调用，这个非常简单，从 unary 中取出 N 个参数，放入调用栈
			wavy := false
			if c.Op1 != nil {
				wavy = c.Op1.Bool()
			}
			args := v.popArgN(c.Unary)
			callableValue := v.pop()
			v.call(callableValue, wavy, args)
		}
	case OpPop:
		// 弹个栈数据出来
		v.lastStackValue = v.pop()
		return
	case OpNewSliceWithType:
		vals := v.popArgN(c.Unary)
		typeLiteral := v.pop()

		var sliceReflectedValue reflect.Value
		if !typeLiteral.IsType() {
			panic("BUG: typeLiteral for typed slice failed, got: " + typeLiteral.String())
		}

		sliceType := typeLiteral.Type()
		if sliceType.Kind() != reflect.Slice {
			panic("BUG: typed slice cannot make type: " + sliceType.String())
		}

		sliceReflectedValue = reflect.MakeSlice(sliceType, c.Unary, c.Unary)
		if c.Unary > 0 {
			for index, val := range vals {
				val := reflect.ValueOf(val.Value)
				err := v.AutoConvertReflectValueByType(&val, sliceType.Elem())
				if err != nil {
					panic(fmt.Sprintf("runtime error: cannot convert `%v` to `%v`", val.String(), sliceType.Elem().String()))
				}
				sliceReflectedValue.Index(index).Set(val)
			}
		}
		v.push(NewAutoValue(sliceReflectedValue.Interface()))
		return
	case OpNewSlice:
		vals := v.popArgN(c.Unary)
		elementType := GuessValuesTypeToBasicType(vals...)
		sliceType := reflect.SliceOf(elementType)
		value := reflect.MakeSlice(sliceType, c.Unary, c.Unary)
		for index, e := range vals {
			val := reflect.ValueOf(e.Value)
			err := v.AutoConvertReflectValueByType(&val, elementType)
			if err != nil {
				panic(fmt.Sprintf("cannot convert %v to %v", val.Type(), elementType))
			}
			value.Index(index).Set(val)
		}
		v.push(NewValue(sliceType.String(), value.Interface(), ""))
		return
	case OpIterableCall:
		isEnd := v.pop()
		args := v.popArgN(c.Unary)
		iterableValue := v.pop()
		argsLength := len(args)

		iterableValueType := reflect.TypeOf(iterableValue.Value)
		var iterableValueRF reflect.Value
		if iterableValueType.Kind() == reflect.String {
			if v, ok := iterableValue.Value.(string); ok {
				iterableValueRF = reflect.ValueOf([]rune(v))
			} else {
				panic("cannot convert string to []byte/rune")
			}
		} else {
			iterableValueRF = reflect.ValueOf(iterableValue.Value)
		}
		if iterableValueRF.Type().Kind() == reflect.Ptr {
			iterableValueRF = iterableValueRF.Elem()
		}
		switch iterableValueRF.Type().Kind() {
		case reflect.String:
			fallthrough
		case reflect.Slice:
			/*
				in nasl
				"123"[0] means "1" instead of '1'
				"123"[111] means nil
			*/
			for _, arg := range args {
				if !arg.IsInt() {
					panic("slice call args must be int")
				}
			}
			var start, end int
			step := 1

			switch argsLength {
			case 3:
				step = args[2].Int()
				if step == 0 {
					panic("slice call step cannot be 0")
				}
				fallthrough
			case 2:
				end = args[1].Int()
				if end < 0 {
					end = iterableValueRF.Len() + end
				}
				if end < 0 || end > iterableValueRF.Len() {
					panic("slice call error, end index out of range")
				}
				if end == 0 && isEnd.True() {
					if step > 0 {
						end = iterableValueRF.Len()
					} else {
						end = -1
					}
				}
				fallthrough
			case 1:
				start = args[0].Int()
				if start < 0 {
					start = iterableValueRF.Len() + start
				}
				if start < 0 || start >= iterableValueRF.Len() {
					switch v.vm.GetConfig().vmMode {
					case NASL:
						v.push(GetUndefined())
						return
					default:
						panic("slice call error, start index out of range")
					}
				}
				if start == 0 && isEnd.True() && step < 0 {
					start = iterableValueRF.Len() - 1
				}
			default:
				panic("slice call error, args got " + fmt.Sprint(argsLength) + ".")
			}
			if step == 0 {
				panic("step cannot be zero")
			}
			var sliceRes reflect.Value
			if argsLength == 1 {
				// 这里可以转变为左值，因为 abc[expr] = xxx 是可以赋值的
				sliceRes = iterableValueRF.Index(start)
				if iterableValueType.Kind() == reflect.String {
					switch v.vm.GetConfig().vmMode {
					case NASL:
						var subStr = string([]rune{sliceRes.Interface().(rune)})
						v.push(NewValue("string", subStr, subStr))
					default:
						v.push(NewValue("char", sliceRes.Interface().(rune), fmt.Sprintf("%c", sliceRes.Interface())))
					}
				} else {
					value := NewValue(sliceRes.Type().String(), sliceRes.Interface(), fmt.Sprint(sliceRes))
					value.CallerRef = iterableValue
					value.CalleeRef = args[0]
					v.push(value)
				}
			} else {
				sliceRes = reflect.MakeSlice(iterableValueRF.Type(), 0, iterableValueRF.Len())

				if step < 0 {
					for i := start; i > end; i += step {
						sliceRes = reflect.Append(sliceRes, iterableValueRF.Index(i))
					}
				} else {
					for i := start; i < end; i += step {
						sliceRes = reflect.Append(sliceRes, iterableValueRF.Index(i))
					}
				}
				if iterableValueType.Kind() == reflect.String {
					if value, ok := sliceRes.Interface().([]rune); ok {
						v.push(NewValue("string", string(value), ""))
					} else {
						panic("cannot convert []byte to string")
					}
				} else {
					v.push(NewValue(iterableValueType.String(), sliceRes.Interface(), ""))
				}
			}
		case reflect.Map:
			if argsLength != 1 {
				panic("map call args must be 1")
			}
			if v.vm.config.vmMode == LUA {
				if args[0].Value == nil {
					v.push(NewValue("nil", nil, ""))
					return
				}
			}
			mapRes := iterableValueRF.MapIndex(reflect.ValueOf(args[0].Value))
			if mapRes.IsValid() {
				v.push(NewValue(mapRes.Type().String(), mapRes.Interface(), ""))
			} else {
				v.push(NewValue("nil", nil, ""))
			}
		case reflect.Struct:
			if argsLength != 1 {
				panic("struct call args length must be 1")
			}
			if !args[0].IsStringOrBytes() {
				panic("struct call args must be string")
			}
			fieldName := args[0].Value.(string)
			memberValue := iterableValueRF.FieldByName(fieldName)
			if !memberValue.IsValid() {
				//panic(fmt.Sprintf("member %s not found in %s", memberNameV.String(), structV.TypeVerbose))
				v.push(undefined)
				return
			} else if !memberValue.CanInterface() {
				v.push(undefined)
				return
				//panic(fmt.Sprintf("member %s is not exported", memberNameV.String()))
			} else {
				//if memberName.Kind() != reflect.Func {
				//	panic(fmt.Sprintf("member %s is not a function", memberNameV.String()))
				//}
				v.push(NewValue(memberValue.String(), memberValue.Interface(), fmt.Sprint(memberValue.Interface())))
				return
			}
		default:
			panic(fmt.Sprintf("'%v' object is not subscriptable", iterableValueRF.Type().String()))
		}
		return
	case OpReturn:
		// 分栈到底还有咩有数据，一般平栈之后，是空的，所以不指定他一定就是 undefined
		if v.stack.Len() > 0 {
			v.lastStackValue = v.pop()
		} else {
			v.lastStackValue = undefined
		}

		// 一般啦
		if v.lastStackValue.IsValueList() {
			v.lastStackValue.Value = v.lastStackValue.ValueListToInterface()
		}
	case OpAssert:
		args := v.popArgN(c.Unary)
		description := "assert error! no description"
		if c.Op1 != nil {
			description = c.Op1.String()
		}
		if len(args) > 1 {
			description = fmt.Sprintf("%v", args[1].Value)
		}
		if args[0].IsBool() {
			if !args[0].Bool() {
				panic(description)
			}
		} else {
			panic("assert args must be bool")
		}
	case OpNewMap:
		switch v.vm.GetConfig().vmMode {
		case LUA:
			var vals []*Value
			if c.Op1 == undefined {
				vals = v.popArgN(2)
				variadic := vals[1].Value
				variadicRF := reflect.ValueOf(variadic)
				index := 0
				values := make([]*Value, 0)
				if variadicRF.IsValid() && variadicRF.Kind() == reflect.Map {
					for index < variadicRF.Len() {
						values = append(values, NewAutoValue(variadicRF.MapIndex(reflect.ValueOf(index+1)).Interface()))
						index++
					}
				} else { // 可变参数接收到空参数 也就是没传参
					values = append(values, undefined)
				}
				mt := reflect.MapOf(literalReflectType_Interface, literalReflectType_Interface)
				mapVal := reflect.MakeMapWithSize(mt, len(values))
				for index := range values {
					vV := reflect.ValueOf(values[index].Value)
					err := v.AutoConvertReflectValueByType(&vV, literalReflectType_Interface)
					if err != nil {
						panic(fmt.Sprintf("cannot convert %v to %v", vV.Type(), "interface{} type"))
					}
					mapVal.SetMapIndex(reflect.ValueOf(index+1), vV)
				}
				v.push(NewValue(mt.String(), mapVal.Interface(), ""))
				return
			} else {
				variadicPresent := false
				variadicPose := 0
				if c.Op1 != nil && c.Op2 == undefined {
					variadicPresent = true
					variadicPose = c.Op1.Int()
				}
				vals = v.popArgN(c.Unary * 2)
				if len(vals) <= 0 {
					v.push(NewGenericMap(""))
					return
				}
				var keys = make([]*Value, c.Unary)
				var values = make([]*Value, c.Unary)
				for i := 0; i < c.Unary; i++ {
					kI := i * 2
					vI := i*2 + 1
					keys[i] = vals[kI]
					values[i] = vals[vI]
				}
				if variadicPresent {
					variadic := values[variadicPose]
					variadicRF := reflect.ValueOf(variadic)
					if variadicRF.IsValid() && variadicRF.Kind() == reflect.Map {
						values[variadicPose] = NewAutoValue(variadicRF.MapIndex(reflect.ValueOf(1)).Interface())
					} else {
						values[variadicPose] = undefined
					}
					if len(values) == 0 { // 可变参数为空map 也就是没传参 那他应该在被用作创建map时表现为nil
						values[0] = undefined
					}
				}
				kType := literalReflectType_Interface
				vType := literalReflectType_Interface
				mt := reflect.MapOf(kType, vType)
				mapVal := reflect.MakeMapWithSize(mt, len(keys))
				for index := range keys {
					kV := reflect.ValueOf(keys[index].Value)
					vV := reflect.ValueOf(values[index].Value)
					err := v.AutoConvertReflectValueByType(&kV, kType)
					if err != nil {
						panic(fmt.Sprintf("cannot convert %v to %v", kV.Type(), kType))
					}
					err = v.AutoConvertReflectValueByType(&vV, vType)
					if err != nil {
						panic(fmt.Sprintf("cannot convert %v to %v", vV.Type(), vType))
					}
					mapVal.SetMapIndex(kV, vV)
				}

				v.push(NewValue(mt.String(), mapVal.Interface(), ""))

				return
			}
		case YAK:
			fallthrough
		default:
			vals := v.popArgN(c.Unary * 2)
			if len(vals) <= 0 {
				v.push(NewEmptyMap(""))
				return
			}
			var keys = make([]*Value, c.Unary)
			var values = make([]*Value, c.Unary)
			for i := 0; i < c.Unary; i++ {
				kI := i * 2
				vI := i*2 + 1
				keys[i] = vals[kI]
				values[i] = vals[vI]
			}
			kType := GuessValuesTypeToBasicType(keys...)
			vType := GuessValuesTypeToBasicType(values...)
			mt := reflect.MapOf(kType, vType)
			mapVal := reflect.MakeMapWithSize(mt, len(keys))
			for index := range keys {
				kV := reflect.ValueOf(keys[index].Value)
				vV := reflect.ValueOf(values[index].Value)
				err := v.AutoConvertReflectValueByType(&kV, kType)
				if err != nil {
					panic(fmt.Sprintf("cannot convert %v to %v", kV.Type(), kType))
				}
				err = v.AutoConvertReflectValueByType(&vV, vType)
				if err != nil {
					panic(fmt.Sprintf("cannot convert %v to %v", vV.Type(), vType))
				}
				mapVal.SetMapIndex(kV, vV)
			}

			v.push(NewValue(mt.String(), mapVal.Interface(), ""))
			return
		}
	case OpNewMapWithType:
		vals := v.popArgN(c.Unary * 2)
		mapTypeLiteral := v.pop()

		if len(vals) <= 0 { // 不应该执行这个if
			v.push(NewEmptyMap(""))
			return
		}
		var keys = make([]*Value, c.Unary)
		var values = make([]*Value, c.Unary)
		for i := 0; i < c.Unary; i++ {
			kI := i * 2
			vI := i*2 + 1
			keys[i] = vals[kI]
			values[i] = vals[vI]
		}
		mapType := mapTypeLiteral.Type()
		kType, vType := mapType.Key(), mapType.Elem()
		mapVal := reflect.MakeMapWithSize(mapType, len(keys))

		for index := range keys {
			kV := reflect.ValueOf(keys[index].Value)
			vV := reflect.ValueOf(values[index].Value)
			err := v.AutoConvertReflectValueByType(&kV, kType)
			if err != nil {
				panic(fmt.Sprintf("cannot convert %v to %v", kV.Type(), kType))
			}
			err = v.AutoConvertReflectValueByType(&vV, vType)
			if err != nil {
				panic(fmt.Sprintf("cannot convert %v to %v", vV.Type(), vType))
			}
			mapVal.SetMapIndex(kV, vV)
		}

		v.push(NewValue(mapType.String(), mapVal.Interface(), ""))
		return
	case OpType:
		switch c.Op1.TypeVerbose {
		case "int":
			v.push(NewType(c.Op1.TypeVerbose, literalReflectType_Int))
		case "byte":
			v.push(NewType(c.Op1.TypeVerbose, literalReflectType_Byte))
		case "string":
			v.push(NewType(c.Op1.TypeVerbose, literalReflectType_String))
		case "bool":
			v.push(NewType(c.Op1.TypeVerbose, literalReflectType_Bool))
		case "float":
			v.push(NewType(c.Op1.TypeVerbose, literalReflectType_Float64))
		case "var":
			v.push(NewType(c.Op1.TypeVerbose, literalReflectType_Interface))
		case "slice":
			if val := v.pop(); val.IsType() {
				v.push(NewType("[]"+val.TypeVerbose, reflect.TypeOf(reflect.MakeSlice(reflect.SliceOf(val.Type()), 0, 0).Interface())))
			}
		case "map":
			if value := v.pop(); value.IsType() {
				if key := v.pop(); key.IsType() {
					v.push(NewType(fmt.Sprintf("map[%s]%s", key.TypeVerbose, value.TypeVerbose), reflect.TypeOf(reflect.MakeMap(reflect.MapOf(key.Type(), value.Type())).Interface())))
				}
			}
		case "chan":
			if val := v.pop(); val.IsType() {
				v.push(NewType("chan "+val.TypeVerbose, reflect.TypeOf(reflect.MakeChan(reflect.ChanOf(reflect.BothDir, val.Type()), 0).Interface())))
			}
		default:
			panic(fmt.Sprintf("invalid type: %s", strconv.Quote(c.Op1.TypeVerbose)))
		}
	case OpMake:
		var newValue interface{}
		var size, makeCap int = 0, 0
		vals := v.popArgN(c.Unary)
		if len(vals) > 0 {
			size = vals[0].Int()
			if len(vals) > 1 {
				makeCap = vals[1].Int()
			}
		}
		if val := v.pop(); val.IsType() {
			t := val.Type()
			switch t.Kind() {
			case reflect.Slice:
				if makeCap <= 0 && size > makeCap {
					makeCap = size
				}
				newValue = reflect.MakeSlice(t, size, makeCap).Interface()
			case reflect.Map:
				newValue = reflect.MakeMapWithSize(t, size).Interface()
				if len(vals) > 1 {
					panic(fmt.Sprintf("make %s expect 1 or 2 arguments, but got 3", val.TypeVerbose))
				}
			case reflect.Chan:
				newValue = reflect.MakeChan(t, size).Interface()
				if len(vals) > 1 {
					panic(fmt.Sprintf("make %s expect 1 or 2 arguments, but got 3", val.TypeVerbose))
				}
			default:
				panic(fmt.Sprintf("cannot make %s", val.TypeVerbose))
			}
			v.push(NewValue(val.TypeVerbose, newValue, fmt.Sprint(newValue)))
			return
		}
	case OpPanic:
		val := v.pop()
		panic(NewVMPanic(val.Value))
	case OpRecover:
		v.push(NewAutoValue(v.recover().GetData()))
	case OpDefer:
		// op defer 应该在整体执行的时候，调用 defer 执行
		_, ok := c.Op1.Value.([]*Code)
		if !ok {
			panic("BUG: defer opcode's code is error!")
		}
	case OpMemberCall:
		/*
			memberCall 最后会产生一个 Value 发送到栈中
				1. 一般来说，这个操作，其实不应该仅仅取出他的值，还应该记录取值路径，字段名是啥？
		*/
		switch v.vm.GetConfig().vmMode {
		case LUA:
			var callerReflectValue reflect.Value

			memberNameV := v.pop()
			//if !memberNameV.IsString() {
			//	panic("member name must be string")
			//}
			memberName := memberNameV.String()
			caller := v.pop()
			//处理map类型
			callerReflectValue = reflect.ValueOf(caller.Value)
			callerTypeKind := callerReflectValue.Kind()
			callerTypeName := callerTypeKind.String()
			if caller.IsBytes() {
				callerTypeName = "bytes"
			}

			switch callerTypeKind {
			case reflect.Map:
				method, ok := mapBuildinMethod[memberName]
				if ok {
					ret := method.HandlerFactory(v, caller)
					literal := fmt.Sprintf("map.%s", memberName)
					v.push(NewValue(literal, ret, literal))
					return
				}

				member := callerReflectValue.MapIndex(reflect.ValueOf(memberNameV.Value))
				if member.IsValid() {
					literal := fmt.Sprintf("%s.%s", caller.Literal, memberName)
					value := NewValue(member.Type().String(), member.Interface(), literal)
					value.CalleeRef = memberNameV
					value.CallerRef = caller
					v.push(value)
					if c.Unary == 0 { // object method
						v.push(caller)
					}
					return
				}

				// map 找不到 key 应该崩掉
				panic(fmt.Sprintf("runtime error: no such key `%v` in map", memberNameV.Value))
			default:

			}
			panic(fmt.Sprintf("cannot find built-in method %s of %s type", memberName, callerTypeName))
			//v.push(undefined)
			return
		case YAK:
			fallthrough
		default:

			var callerReflectValue reflect.Value

			memberNameV := v.pop()
			//if !memberNameV.IsString() {
			//	panic("member name must be string")
			//}
			memberName := memberNameV.String()
			caller := v.pop()
			//处理map类型
			callerReflectValue = reflect.ValueOf(caller.Value)
			callerTypeKind := callerReflectValue.Kind()
			callerTypeName := callerTypeKind.String()
			if caller.IsBytes() {
				callerTypeName = "bytes"
			}

			switch callerTypeKind {
			case reflect.Map:

				// 先尝试获取 map 的成员
				getMember := func() (v reflect.Value) {
					defer func() {
						if e := recover(); e != nil {

						}
					}()
					return callerReflectValue.MapIndex(reflect.ValueOf(memberNameV.Value))
				}
				member := getMember()
				if member.IsValid() {
					literal := fmt.Sprintf("%s.%s", caller.Literal, memberName)

					calleeValue := member.Interface()
					calleeValue = v.execHijackMapMemberCallHandler(caller.Literal, memberName, calleeValue)
					value := NewValue(member.Type().String(), calleeValue, literal)
					value.CalleeRef = memberNameV
					value.CallerRef = caller
					v.push(value)
					return
				}
				if caller.GetExtraInfo("global") == nil { // 如果是普通变量的话，尝试获取内置方法
					// 尝试获取内置方法
					method, ok := mapBuildinMethod[memberName]
					if ok {
						ret := method.HandlerFactory(v, caller)
						literal := fmt.Sprintf("map.%s", memberName)
						v.push(NewValue(literal, ret, literal))
						return
					}
				}
				// map 找不到 key 应该崩掉
				panic(fmt.Sprintf("runtime error: no such key `%v` in map", memberNameV.Value))

			case reflect.Ptr, reflect.Struct:
				firstChar := memberName[0]
				if firstChar >= 'a' && firstChar <= 'z' {
					memberName = strings.ToUpper(string(firstChar)) + memberName[1:]
				}
				fun := callerReflectValue.MethodByName(memberName)
				if fun.IsValid() {
					literal := fmt.Sprintf("%s.%s", caller.Literal, memberName)
					value := NewValue(fun.Type().String(), fun.Interface(), literal)
					value.CalleeRef = memberNameV
					value.CallerRef = caller
					v.push(value)
					return
				}
				if callerTypeKind == reflect.Ptr {
					callerReflectValue = callerReflectValue.Elem()
					fun = callerReflectValue.MethodByName(memberName)
					if fun.IsValid() {
						literal := fmt.Sprintf("%s.%s", caller.Literal, memberName)
						value := NewValue(fun.Type().String(), fun.Interface(), literal)
						value.CalleeRef = memberNameV
						value.CallerRef = caller
						v.push(value)
						return
					}
				}

				//获取结构体字段
				if callerReflectValue.Kind() == reflect.Struct {
					member := callerReflectValue.FieldByName(memberName)
					if !member.IsValid() {
						panicByNoSuchKey(memberName, caller.Value)
						return
					}
					if member.CanInterface() {
						literal := fmt.Sprintf("%s.%s", caller.Literal, memberName)
						value := NewValue(member.Type().String(), member.Interface(), literal)
						value.CalleeRef = memberNameV
						value.CallerRef = caller
						v.push(value)
						return
					} else {
						panicByNoSuchKey(memberName, caller.Value)
						return
					}
				}
			case reflect.Array, reflect.Slice:
				//memberName = _title(memberName)
				prefix := "array"
				targetBuildinMethod := arrayBuildinMethod

				if caller.IsBytes() {
					prefix = "bytes"
					targetBuildinMethod = bytesBuildinMethod
				}

				method, ok := targetBuildinMethod[memberName]
				if ok {
					ret := method.HandlerFactory(v, caller)
					literal := fmt.Sprintf("%s.%s", prefix, memberName)
					v.push(NewValue(literal, ret, literal))
					return
				}
			case reflect.String:
				//memberName = _title(memberName)
				method, ok := stringBuildinMethod[memberName]

				if ok {
					literal := fmt.Sprintf("string.%s", memberName)
					v.push(NewValue(literal, method.HandlerFactory(v, caller.Value), literal))
					return
				}

			default:

			}
			panic(fmt.Sprintf("cannot find built-in method %s of %s type", memberName, callerTypeName))
			//v.push(undefined)
			return
		}
	case OpEllipsis:
		//AutoConvertReflectValueByType()
		arrayRaw := v.pop()
		code := v.peekNextCode()

		if arrayRaw.IsBytes() {
			d := arrayRaw.Bytes()
			for _, r := range d {
				v.push(NewValue("byte", r, fmt.Sprintf("%c", r)))
			}
			code.Unary = c.Unary + len(d) - 1
			return
		}
		if arrayRaw.IsString() {
			d := []rune(arrayRaw.String())
			for _, r := range d {
				v.push(NewValue("char", r, fmt.Sprintf("%c", r)))
			}
			code.Unary = c.Unary + len(d) - 1
			return
		}
		if reflect.TypeOf(arrayRaw.Value).Kind() == reflect.Slice {
			refV := reflect.ValueOf(arrayRaw.Value)
			for i := 0; i < refV.Len(); i++ {
				v.push(NewValue("char", refV.Index(i).Interface(), fmt.Sprint(refV.Index(i).Interface())))
			}
			code.Unary = c.Unary + refV.Len() - 1
		}
	case OpExit:
		val := v.pop()
		panic(NewVMPanic(&VMPanicSignal{Info: val}))
	}
}

func (v *Frame) execOp1(op OpcodeFlag, op1 *Value) *Value {
	h, ok := v.UnaryOperatorTable[op]
	if !ok {
		log.Errorf("cannot support unary op: %v", OpcodeToName(op))
		return undefined
	}
	return h(op1)
}

func (v *Frame) execOp2(op OpcodeFlag, op1, op2 *Value) *Value {
	h, ok := v.BinaryOperatorTable[op]
	if !ok {
		log.Errorf("cannot support binary op: %v", OpcodeToName(op))
		return undefined
	}
	return h(op1, op2)
}

func panicByNoSuchKey(mem string, i interface{}) {
	structHelper, _ := yakdocument.Dir(i)
	if structHelper != nil {
		var fields []string
		var methods []string
		for _, m := range structHelper.Methods {
			methods = append(methods, m.Name)
		}
		for _, m := range structHelper.PtrMethods {
			methods = append(methods, m.Name)
		}
		for _, f := range structHelper.Fields {
			fields = append(fields, f.Name)
		}
		if len(methods) > 0 || len(fields) > 0 {
			var buf bytes.Buffer
			if len(methods) > 0 && len(fields) > 0 {
				buf.WriteString(fmt.Sprintf("fields:%v or methods:%v", spew.Sdump(fields), spew.Sdump(utils.RemoveRepeatStringSlice(methods))))
			} else if len(methods) > 0 && fields == nil {
				buf.WriteString(fmt.Sprintf("methods:%v", spew.Sdump(methods)))
			} else {
				buf.WriteString(fmt.Sprintf("fields:%v", spew.Sdump(fields)))
			}
			panic("no such field or method: " + mem +
				" in native struct or ptr native struct!\n do you mean one of " + buf.String())
		}
	}
	panic("no such field or method: " + mem + " in native struct or ptr native struct!")
}

/*
push int 1
push int 2
op add
pop1

binaryOp:
	pop arg 2
		---exec(opSig, op1, op2)
	push ret

*/