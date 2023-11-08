package types

import fmt "fmt"

// WrapMsg takes a PrefixMsgType and a message byte slice and returns a new byte slice
// with the appropriate prefix added based on the message type.
func WrapMsg(prefixMsgType PrefixMsgType, msg []byte) []byte {
	switch prefixMsgType {
	case PREFIX_MSG_TYPE_TEXT:
		return wrapTextMsgNormal(msg)
	case PREFIX_MSG_TYPE_REPLACE_GROUP:
		return wrapReplaceGroupMsg(msg)
	case PREFIX_MSG_TYPE_ORACLE:
		return wrapOracleMsg(msg)
	default:
		panic(fmt.Errorf("message PrefixMsgType does not support %s type", prefixMsgType))
	}
}

// wrapTextMsgNormal appends the text message prefix to the given message bytes.
func wrapTextMsgNormal(msg []byte) []byte {
	return append(NormalMsgPrefix, msg...)
}

// wrapReplaceGroupMsg constructs a message by appending the replace group message prefix,
// the public key, and the formatted time to the message bytes.
func wrapReplaceGroupMsg(pubKey []byte) []byte {
	return append(ReplaceGroupMsgPrefix, pubKey...)
}

// wrapOracleMsg appends the oracle result message prefix to the given oracle result bytes.
func wrapOracleMsg(result []byte) []byte {
	return append(OracleMsgPrefix, result...)
}
