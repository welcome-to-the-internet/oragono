package irc

import (
	"fmt"
	"strings"
	"time"
)

type Identifier interface {
	Id() string
	PublicId() string
	Nick() string
}

type Replier interface {
	Replies() chan<- Reply
}

type Reply interface {
	Format(client *Client) string
	Source() Identifier
}

type BaseReply struct {
	source  Identifier
	message string
}

func (reply *BaseReply) Source() Identifier {
	return reply.source
}

type StringReply struct {
	BaseReply
	code string
}

func NewStringReply(source Identifier, code string,
	format string, args ...interface{}) *StringReply {
	message := fmt.Sprintf(format, args...)
	fullMessage := fmt.Sprintf(":%s %s %s", source.Id(), code, message)
	return &StringReply{
		BaseReply: BaseReply{source, fullMessage},
		code:      code,
	}
}

func (reply *StringReply) Format(client *Client) string {
	return reply.message
}

func (reply *StringReply) String() string {
	return fmt.Sprintf("Reply(source=%s, code=%s, message=%s)",
		reply.source, reply.code, reply.message)
}

type NumericReply struct {
	BaseReply
	code int
}

func NewNumericReply(source Identifier, code int, format string,
	args ...interface{}) *NumericReply {
	return &NumericReply{
		BaseReply: BaseReply{source, fmt.Sprintf(format, args...)},
		code:      code,
	}
}

func (reply *NumericReply) Format(client *Client) string {
	return fmt.Sprintf(":%s %03d %s %s", reply.Source().Id(), reply.code,
		client.Nick(), reply.message)
}

func (reply *NumericReply) String() string {
	return fmt.Sprintf("Reply(source=%s, code=%d, message=%s)",
		reply.source, reply.code, reply.message)
}

// messaging replies

func RplPrivMsg(source Identifier, target Identifier, message string) Reply {
	return NewStringReply(source, RPL_PRIVMSG, "%s :%s", target.Nick(), message)
}

func RplNick(source Identifier, newNick string) Reply {
	return NewStringReply(source, RPL_NICK, newNick)
}

func RplPrivMsgChannel(channel *Channel, source Identifier, message string) Reply {
	return NewStringReply(source, RPL_PRIVMSG, "%s :%s", channel.name, message)
}

func RplJoin(channel *Channel, user *User) Reply {
	return NewStringReply(user, RPL_JOIN, channel.name)
}

func RplPart(channel *Channel, user *User, message string) Reply {
	return NewStringReply(user, RPL_PART, "%s :%s", channel.name, message)
}

func RplPong(server *Server) Reply {
	return NewStringReply(server, RPL_PONG, server.Id())
}

func RplQuit(client *Client, message string) Reply {
	return NewStringReply(client, RPL_QUIT, ":%s", message)
}

func RplInviteMsg(channel *Channel, inviter *Client) Reply {
	return NewStringReply(inviter, RPL_INVITE, channel.name)
}

// numeric replies

func RplWelcome(source Identifier, client *Client) Reply {
	return NewNumericReply(source, RPL_WELCOME,
		"Welcome to the Internet Relay Network %s", client.Id())
}

func RplYourHost(server *Server, target *Client) Reply {
	return NewNumericReply(server, RPL_YOURHOST,
		"Your host is %s, running version %s", server.name, VERSION)
}

func RplCreated(server *Server) Reply {
	return NewNumericReply(server, RPL_CREATED,
		"This server was created %s", server.ctime.Format(time.RFC1123))
}

func RplMyInfo(server *Server) Reply {
	return NewNumericReply(server, RPL_MYINFO,
		"%s %s a kn", server.name, VERSION)
}

func RplUModeIs(server *Server, client *Client) Reply {
	return NewNumericReply(server, RPL_UMODEIS, client.UModeString())
}

func RplNoTopic(channel *Channel) Reply {
	return NewNumericReply(channel.server, RPL_NOTOPIC,
		"%s :No topic is set", channel.name)
}

func RplTopic(channel *Channel) Reply {
	return NewNumericReply(channel.server, RPL_TOPIC,
		"%s :%s", channel.name, channel.topic)
}

func RplInvitingMsg(channel *Channel, invitee *Client) Reply {
	return NewNumericReply(channel.server, RPL_INVITING,
		"%s %s", channel.name, invitee.Nick())
}

func RplNamReply(channel *Channel) Reply {
	// TODO multiple names and splitting based on message size
	return NewNumericReply(channel.server, RPL_NAMREPLY,
		"= %s :%s", channel.name, strings.Join(channel.Nicks(), " "))
}

func RplEndOfNames(source Identifier) Reply {
	return NewNumericReply(source, RPL_ENDOFNAMES,
		":End of NAMES list")
}

func RplYoureOper(server *Server) Reply {
	return NewNumericReply(server, RPL_YOUREOPER,
		":You are now an IRC operator")
}

// errors (also numeric)

func ErrAlreadyRegistered(source Identifier) Reply {
	return NewNumericReply(source, ERR_ALREADYREGISTRED,
		":You may not reregister")
}

func ErrNickNameInUse(source Identifier, nick string) Reply {
	return NewNumericReply(source, ERR_NICKNAMEINUSE,
		"%s :Nickname is already in use", nick)
}

func ErrUnknownCommand(source Identifier, command string) Reply {
	return NewNumericReply(source, ERR_UNKNOWNCOMMAND,
		"%s :Unknown command", command)
}

func ErrUsersDontMatch(source Identifier) Reply {
	return NewNumericReply(source, ERR_USERSDONTMATCH,
		":Cannot change mode for other users")
}

func ErrNeedMoreParams(source Identifier, command string) Reply {
	return NewNumericReply(source, ERR_NEEDMOREPARAMS,
		"%s :Not enough parameters", command)
}

func ErrNoSuchChannel(source Identifier, channel string) Reply {
	return NewNumericReply(source, ERR_NOSUCHCHANNEL,
		"%s :No such channel", channel)
}

func ErrUserOnChannel(channel *Channel, member *Client) Reply {
	return NewNumericReply(channel.server, ERR_USERONCHANNEL,
		"%s %s :is already on channel", member.nick, channel.name)
}

func ErrNotOnChannel(channel *Channel) Reply {
	return NewNumericReply(channel.server, ERR_NOTONCHANNEL,
		"%s :You're not on that channel", channel.name)
}

func ErrInviteOnlyChannel(channel *Channel) Reply {
	return NewNumericReply(channel.server, ERR_INVITEONLYCHAN,
		"%s :Cannot join channel (+i)", channel.name)
}

func ErrBadChannelKey(channel *Channel) Reply {
	return NewNumericReply(channel.server, ERR_BADCHANNELKEY,
		"%s :Cannot join channel (+k)", channel.name)
}

func ErrNoSuchNick(source Identifier, nick string) Reply {
	return NewNumericReply(source, ERR_NOSUCHNICK,
		"%s :No such nick/channel", nick)
}

func ErrPasswdMismatch(server *Server) Reply {
	return NewNumericReply(server, ERR_PASSWDMISMATCH, ":Password incorrect")
}

func ErrNoChanModes(channel *Channel) Reply {
	return NewNumericReply(channel.server, ERR_NOCHANMODES,
		"%s :Channel doesn't support modes", channel.name)
}

func ErrNoPrivileges(server *Server) Reply {
	return NewNumericReply(server, ERR_NOPRIVILEGES, ":Permission Denied")
}

func ErrRestricted(server *Server) Reply {
	return NewNumericReply(server, ERR_RESTRICTED, ":Your connection is restricted!")
}