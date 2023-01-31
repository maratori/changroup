// Package changroup implements publish/subscribe pattern with group of channels.
//
// A value is sent to each channel in the group. Channels can be acquired/released dynamically.
//
// [Group] allows to acquire/release channel and to send a value to all acquired channels.
//
// [AckableGroup] does the same, but sends [Ackable] value.
// It calls original ack function only after all subscribers acked their copy of value.
// It's useful if you need to know when the message is processed.
package changroup
