package message

import (
	"fmt"
)

type messageImpl struct {
	payload    []byte
	properties propertiesImpl
}

// MessageType returns the type of message.
func (m *messageImpl) MessageType() MessageType {
	val, ok := m.properties.Get(messageTypeKey)
	if !ok {
		return MessageTypeUnknown
	}
	return unmarshalMessageType(val)
}

// Version returns the message format version.
func (m *messageImpl) Version() Version {
	value, ok := m.properties.Get(messageVersion)
	if !ok {
		return VersionOld
	}
	return newMessageVersionFromString(value)
}

// Payload returns payload of current message.
func (m *messageImpl) Payload() []byte {
	return m.payload
}

// Properties returns the message properties.
func (m *messageImpl) Properties() RProperties {
	return m.properties
}

// EstimateSize returns the estimated size of current message.
func (m *messageImpl) EstimateSize() int {
	// TODO: more accurate size estimation.
	return len(m.payload) + m.properties.EstimateSize()
}

// WithVChannel sets the virtual channel of current message.
func (m *messageImpl) WithVChannel(vChannel string) MutableMessage {
	if m.properties.Exist(messageVChannel) {
		panic("vchannel already set in properties of message")
	}
	m.properties.Set(messageVChannel, vChannel)
	return m
}

// WithTimeTick sets the time tick of current message.
func (m *messageImpl) WithTimeTick(tt uint64) MutableMessage {
	if m.properties.Exist(messageTimeTick) {
		panic("time tick already set in properties of message")
	}
	m.properties.Set(messageTimeTick, EncodeUint64(tt))
	return m
}

// WithLastConfirmed sets the last confirmed message id of current message.
func (m *messageImpl) WithLastConfirmed(id MessageID) MutableMessage {
	if m.properties.Exist(messageLastConfirmed) {
		panic("last confirmed message already set in properties of message")
	}
	m.properties.Set(messageLastConfirmed, id.Marshal())
	return m
}

// WithLastConfirmedUseMessageID sets the last confirmed message id of current message to be the same as message id.
func (m *messageImpl) WithLastConfirmedUseMessageID() MutableMessage {
	m.properties.Set(messageLastConfirmed, messageLastConfirmedValueUseMessageID)
	return m
}

// IntoImmutableMessage converts current message to immutable message.
func (m *messageImpl) IntoImmutableMessage(id MessageID) ImmutableMessage {
	return &immutableMessageImpl{
		messageImpl: *m,
		id:          id,
	}
}

// TimeTick returns the time tick of current message.
func (m *messageImpl) TimeTick() uint64 {
	value, ok := m.properties.Get(messageTimeTick)
	if !ok {
		panic("there's a bug in the message codes, timetick lost in properties of message")
	}
	tt, err := DecodeUint64(value)
	if err != nil {
		panic(fmt.Sprintf("there's a bug in the message codes, dirty timetick %s in properties of message", value))
	}
	return tt
}

// VChannel returns the vchannel of current message.
// If the message is broadcasted, the vchannel will be empty.
func (m *messageImpl) VChannel() string {
	value, ok := m.properties.Get(messageVChannel)
	if !ok {
		return ""
	}
	return value
}

type immutableMessageImpl struct {
	messageImpl
	id MessageID
}

// WALName returns the name of message related wal.
func (m *immutableMessageImpl) WALName() string {
	return m.id.WALName()
}

// MessageID returns the message id.
func (m *immutableMessageImpl) MessageID() MessageID {
	return m.id
}

func (m *immutableMessageImpl) LastConfirmedMessageID() MessageID {
	value, ok := m.properties.Get(messageLastConfirmed)
	if !ok {
		panic(fmt.Sprintf("there's a bug in the message codes, last confirmed message lost in properties of message, id: %+v", m.id))
	}
	if value == messageLastConfirmedValueUseMessageID {
		return m.MessageID()
	}
	id, err := UnmarshalMessageID(m.id.WALName(), value)
	if err != nil {
		panic(fmt.Sprintf("there's a bug in the message codes, dirty last confirmed message in properties of message, id: %+v", m.id))
	}
	return id
}
