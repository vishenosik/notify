package rabbitmq

import "fmt"

func (ac *AmqpClient) NewPublisherChannel(conf *PublisherConfig) (*Channel, error) {
	channel, err := ac.newChannel(conf)
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(
		conf.Exchange,
		exchangeKindDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	if conf.Queue != "" {
		if _, err := channel.QueueDeclare(
			conf.Queue,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("declare queue: %w", err)
		}
		if err := channel.QueueBind(
			conf.Queue,
			conf.RoutingKey,
			conf.Exchange,
			false,
			nil,
		); err != nil {
			return nil, fmt.Errorf("bind queue: %w", err)
		}
	}

	return channel, nil
}
