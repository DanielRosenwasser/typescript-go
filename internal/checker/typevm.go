package checker

type typeInstantiationProgramKind uint8

const (
	typeInstantiationProgramNone typeInstantiationProgramKind = iota
	typeInstantiationProgramTypeParameter
	typeInstantiationProgramIndex
	typeInstantiationProgramTemplateLiteral
	typeInstantiationProgramStringMapping
)

type typeInstantiationOp uint8

const (
	typeInstantiationOpPushCurrentType typeInstantiationOp = iota
	typeInstantiationOpPushCurrentMapper
	typeInstantiationOpPushCurrentTemplateTypes
	typeInstantiationOpMapType
	typeInstantiationOpGetTarget
	typeInstantiationOpInstantiateType
	typeInstantiationOpInstantiateTypeList
	typeInstantiationOpGetIndexType
	typeInstantiationOpGetStringMappingTypeFromCurrent
	typeInstantiationOpGetTemplateLiteralTypeFromCurrent
	typeInstantiationOpReturnType
)

type typeInstantiationInstr struct {
	op typeInstantiationOp
}

var typeInstantiationPrograms = [...][]typeInstantiationInstr{
	typeInstantiationProgramTypeParameter: {
		{op: typeInstantiationOpPushCurrentType},
		{op: typeInstantiationOpPushCurrentMapper},
		{op: typeInstantiationOpMapType},
		{op: typeInstantiationOpReturnType},
	},
	typeInstantiationProgramIndex: {
		{op: typeInstantiationOpPushCurrentType},
		{op: typeInstantiationOpGetTarget},
		{op: typeInstantiationOpPushCurrentMapper},
		{op: typeInstantiationOpInstantiateType},
		{op: typeInstantiationOpGetIndexType},
		{op: typeInstantiationOpReturnType},
	},
	typeInstantiationProgramTemplateLiteral: {
		{op: typeInstantiationOpPushCurrentTemplateTypes},
		{op: typeInstantiationOpPushCurrentMapper},
		{op: typeInstantiationOpInstantiateTypeList},
		{op: typeInstantiationOpGetTemplateLiteralTypeFromCurrent},
		{op: typeInstantiationOpReturnType},
	},
	typeInstantiationProgramStringMapping: {
		{op: typeInstantiationOpPushCurrentType},
		{op: typeInstantiationOpGetTarget},
		{op: typeInstantiationOpPushCurrentMapper},
		{op: typeInstantiationOpInstantiateType},
		{op: typeInstantiationOpGetStringMappingTypeFromCurrent},
		{op: typeInstantiationOpReturnType},
	},
}

type typeInstantiationVM struct {
	c       *Checker
	t       *Type
	mapper  *TypeMapper
	types   [4]*Type
	mappers [2]*TypeMapper
	lists   [2][]*Type
	topType int
	topMap  int
	topList int
}

func (c *Checker) tryRunTypeInstantiationVM(t *Type, m *TypeMapper) (*Type, bool) {
	programKind := getTypeInstantiationProgramKind(t)
	if programKind == typeInstantiationProgramNone {
		return nil, false
	}
	vm := typeInstantiationVM{c: c, t: t, mapper: m}
	return vm.run(typeInstantiationPrograms[programKind]), true
}

func getTypeInstantiationProgramKind(t *Type) typeInstantiationProgramKind {
	flags := t.flags
	switch {
	case flags&TypeFlagsTypeParameter != 0:
		return typeInstantiationProgramTypeParameter
	case flags&TypeFlagsIndex != 0:
		return typeInstantiationProgramIndex
	case flags&TypeFlagsTemplateLiteral != 0:
		return typeInstantiationProgramTemplateLiteral
	case flags&TypeFlagsStringMapping != 0:
		return typeInstantiationProgramStringMapping
	}
	return typeInstantiationProgramNone
}

func (vm *typeInstantiationVM) run(program []typeInstantiationInstr) *Type {
	for _, instr := range program {
		switch instr.op {
		case typeInstantiationOpPushCurrentType:
			vm.pushType(vm.t)
		case typeInstantiationOpPushCurrentMapper:
			vm.pushMapper(vm.mapper)
		case typeInstantiationOpPushCurrentTemplateTypes:
			vm.pushList(vm.t.AsTemplateLiteralType().types)
		case typeInstantiationOpMapType:
			mapper := vm.popMapper()
			t := vm.popType()
			vm.pushType(mapper.Map(t))
		case typeInstantiationOpGetTarget:
			vm.pushType(vm.popType().Target())
		case typeInstantiationOpInstantiateType:
			mapper := vm.popMapper()
			t := vm.popType()
			vm.pushType(vm.c.instantiateType(t, mapper))
		case typeInstantiationOpInstantiateTypeList:
			mapper := vm.popMapper()
			types := vm.popList()
			vm.pushList(vm.c.instantiateTypes(types, mapper))
		case typeInstantiationOpGetIndexType:
			vm.pushType(vm.c.getIndexType(vm.popType()))
		case typeInstantiationOpGetStringMappingTypeFromCurrent:
			vm.pushType(vm.c.getStringMappingType(vm.t.symbol, vm.popType()))
		case typeInstantiationOpGetTemplateLiteralTypeFromCurrent:
			vm.pushType(vm.c.getTemplateLiteralType(vm.t.AsTemplateLiteralType().texts, vm.popList()))
		case typeInstantiationOpReturnType:
			return vm.popType()
		}
	}
	panic("type instantiation program completed without return")
}

func (vm *typeInstantiationVM) pushType(t *Type) {
	vm.types[vm.topType] = t
	vm.topType++
}

func (vm *typeInstantiationVM) popType() *Type {
	vm.topType--
	t := vm.types[vm.topType]
	vm.types[vm.topType] = nil
	return t
}

func (vm *typeInstantiationVM) pushMapper(mapper *TypeMapper) {
	vm.mappers[vm.topMap] = mapper
	vm.topMap++
}

func (vm *typeInstantiationVM) popMapper() *TypeMapper {
	vm.topMap--
	mapper := vm.mappers[vm.topMap]
	vm.mappers[vm.topMap] = nil
	return mapper
}

func (vm *typeInstantiationVM) pushList(types []*Type) {
	vm.lists[vm.topList] = types
	vm.topList++
}

func (vm *typeInstantiationVM) popList() []*Type {
	vm.topList--
	types := vm.lists[vm.topList]
	vm.lists[vm.topList] = nil
	return types
}
