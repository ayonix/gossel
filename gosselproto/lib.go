package gosselproto

func AuthMsg(user, pw string) *Msg {
	return &Msg{
		MessageType: &Msg_Auth{
			Auth: &Auth{
				Username: user,
				Password: pw,
			},
		},
	}
}

func SubscribeMsg(network, channel string, unsubscribe bool) *Msg {
	return &Msg{
		MessageType: &Msg_Subscribe{
			Subscribe: &Subscribe{
				Network:     network,
				Channel:     channel,
				Unsubscribe: unsubscribe,
			},
		},
	}
}

func IdentityMsg(user, real, away string) *Identity {
	return &Identity{
		Username: user,
		Realname: real,
		Awaymsg:  away,
	}
}

func NetworkMsg(name, host, password string, tls bool, identity *Identity, add bool) *Msg {
	return &Msg{
		MessageType: &Msg_Network{
			Network: &Network{
				Name:     name,
				Network:  host,
				Password: password,
				Tls:      tls,
				Identity: identity,
				Add:      add,
			},
		},
	}
}

func Prefix(name, user, host string) *Irc_Prefix {
	return &Irc_Prefix{
		Name: name,
		User: user,
		Host: host,
	}
}

func IrcMsg(network, command string, prefix *Irc_Prefix, params []string, trailing string) *Msg {
	return &Msg{
		MessageType: &Msg_Irc{
			Irc: &Irc{
				Command:  command,
				Prefix:   prefix,
				Params:   params,
				Trailing: trailing,
				Network:  network,
			},
		},
	}
}
