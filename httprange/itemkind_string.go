// Code generated by "stringer -type=itemKind lex.go"; DO NOT EDIT

package httprange

import "fmt"

const _itemKind_name = "itemErroritemEOFitemUnititemStartitemEnditemLength"

var _itemKind_index = [...]uint8{0, 9, 16, 24, 33, 40, 50}

func (i itemKind) String() string {
	if i < 0 || i >= itemKind(len(_itemKind_index)-1) {
		return fmt.Sprintf("itemKind(%d)", i)
	}
	return _itemKind_name[_itemKind_index[i]:_itemKind_index[i+1]]
}
