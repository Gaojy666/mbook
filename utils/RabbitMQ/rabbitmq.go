package RabbitMQ

import (
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/streadway/amqp"
	"log"
	"ziyoubiancheng/mbook/models"
)

// url 格式： amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
const MQURL = "amqp://luxiayuai:lu741208@127.0.0.1:5672/imooc"

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机
	Exchange string
	// key
	Key string
	// 连接信息
	Mqurl string
}

// 创建RabbitMQ结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl:     MQURL,
	}
	var err error
	// 创建rabbitmq连接
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建连接错误！")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败！")
	return rabbitmq
}

// 断开channel和connection
func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

// 错误处理
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// 简单模式step1:创建simple模式的实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	// simple模式使用rabbitmq默认的交换机，key就是没有
	return NewRabbitMQ(queueName, "", "")
}

// 订阅模式创建RabbitMQ实例
func NewRabbitMQPubSub(exchangeName string) *RabbitMQ {
	// 创建RabbitMQ实例
	rabbitmq := NewRabbitMQ("", exchangeName, "")
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建连接错误！")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败！")
	return rabbitmq
}

// 简单模式step2：简单模式下生产代码
func (r *RabbitMQ) PublishSimple(message string) error {
	// 1.申请队列，如果队列不存在，会自动创建；如果创建，则跳过
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 控制消息是否持久化
		false,
		// 是否为自动删除
		false,
		// 是否具有排他性，不常用
		false,
		// 是否阻塞,false为阻塞等待
		false,
		// 额外属性
		nil,
	)
	if err != nil {
		logs.Error(err)
	}
	// 2.发送消息到队列中
	err = r.channel.Publish(
		r.Exchange,
		r.QueueName,
		// mandatory如果为true，会根据exchange类型，和routkey规则，
		// 如果无法找到符合条件的队列，
		// 那么会将发送的消息返回给生产者
		false,
		// immediate如果为true，当exchange发送消息到队列后，
		// 发现队列上没有绑定消费者
		// 则会把消息发还给生产者
		false,
		// 消息
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		logs.Error(err)
	}
	return err
}

//func (r *RabbitMQ) ConsumeSimple() {
//	// 1.申请队列，如果队列不存在，会自动创建；如果创建，则跳过
//	// 保证队列存在，消息能发送到队列中
//	_, err := r.channel.QueueDeclare(
//		r.QueueName,
//		// 控制消息是否持久化
//		false,
//		// 是否为自动删除
//		false,
//		// 是否具有排他性，不常用,即队列是否只有自己可见
//		false,
//		// 是否阻塞,false为阻塞等待
//		false,
//		// 额外属性
//		nil,
//	)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	// 2.接收消息
//	msgs, err := r.channel.Consume(
//		r.QueueName,
//		// 用来区分多个消费者
//		"",
//		// 是否自动应答
//		true,
//		// 是否具有排他性，不常用,即队列是否只有自己可见
//		false,
//		// 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
//		false,
//		// 是否阻塞,false为阻塞等待
//		false,
//		nil,
//	)
//
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	forever := make(chan bool)
//	// 3.启用协程处理消息
//	go func() {
//		for d := range msgs {
//			// 实现我们要处理的逻辑函数
//			log.Printf("Received a message: %s", d.Body)
//		}
//	}()
//	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
//	// 阻塞一下
//	<-forever
//}

func (r *RabbitMQ) ConsumeSimple() error {
	// 1.申请队列，如果队列不存在，会自动创建；如果创建，则跳过
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 控制消息是否持久化
		false,
		// 是否为自动删除
		false,
		// 是否具有排他性，不常用,即队列是否只有自己可见
		false,
		// 是否阻塞,false为阻塞等待
		false,
		// 额外属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// 消费者流控，防止爆库
	// 防止消费者处理消息的速度过快，导致数据库（或其他后端存储）的负载过高，甚至引发系统崩溃。
	r.channel.Qos(
		1,     // 当前消费者从rabbitmq中只接受一个消息
		0,     // 服务器传递的最大容量（以八位字节为单位）没有限制。设为0意味着消费者可以一次性从 RabbitMQ 中获取尽可能多的消息，而不受限制。
		false, // 设置为true，则整个channel全局可用，其他的消费队列也可用
	)

	// 2.接收消息
	msgs, err := r.channel.Consume(
		r.QueueName,
		// 用来区分多个消费者
		"",
		// 当 autoAck 设置为 true 时，消费者在接收到消息后立即自动确认，而无需等待消息处理完成。
		false,
		// 是否具有排他性，不常用,即队列是否只有自己可见
		false,
		// 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		// 是否阻塞,false为阻塞等待
		false,
		nil,
	)

	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)
	// 3.启用协程处理消息
	go func() {
		for d := range msgs {
			// 消息逻辑处理，可自行设计逻辑
			log.Printf("Received a message: %s", d.Body)

			message := &models.Message{}
			err := json.Unmarshal([]byte(d.Body), message)
			if err != nil {
				logs.Error(err)
			}
			models.NewDocument().ReleaseContent(message.BookId, message.BaseUrl)
			models.ElasticBuildIndex(message.BookId)

			// 如果为true，表示确认所有未确认的消息，
			// 为false，表示确认当前消息
			// 也就是说，只有当前业务处理完了后，才会向消息队列发送ack
			d.Ack(false)
		}
	}()
	log.Printf("[*] Waiting for messages, To exit press CTRL+C")
	// 阻塞一下
	<-forever
	return nil
}

// 订阅模式生产
func (r *RabbitMQ) PublishPub(message string) {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型，fanout广播类型
		"fanout",
		true,
		false,
		// internal为true表示这个exchange不可以被client用来推送消息，仅用来进行
		// exchange和exchange之间的绑定
		false,
		false,
		nil,
	)

	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	err = r.channel.Publish(
		r.Exchange,
		// 不指定队列
		"",
		false,
		false,
		// 消息
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// 订阅模式消费端代码
func (r *RabbitMQ) RecieveSub() {
	// 1.试探性创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型，fanout广播类型
		"fanout",
		true,
		false,
		// internal为true表示这个exchange不可以被client用来推送消息，仅用来进行
		// exchange和exchange之间的绑定
		false,
		false,
		nil,
	)

	r.failOnErr(err, "Failed to declare an exchange")

	// 2.试探性创建队列，这里注意队列名称不要写
	q, err := r.channel.QueueDeclare(
		"", // 随机生成队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	// 3.绑定队列到 exchange中
	err = r.channel.QueueBind(
		q.Name,
		// 在pub/sub模式下，这里的key要为空,即随机绑定交换机
		"",
		r.Exchange,
		false,
		nil,
	)

	// 消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C\n")
	<-forever
}

// 路由模式发送消息
func (r *RabbitMQ) PublishRouting(message string) {
	// 1. 尝试创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型，fanout广播类型
		"direct",
		true,
		false,
		// internal为true表示这个exchange不可以被client用来推送消息，仅用来进行
		// exchange和exchange之间的绑定
		false,
		false,
		nil,
	)

	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	err = r.channel.Publish(
		r.Exchange,
		// 不指定队列
		r.Key,
		false,
		false,
		// 消息
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

// 路由模式消费端代码
func (r *RabbitMQ) RecieveRouting() {
	// 1.试探性创建交换机
	err := r.channel.ExchangeDeclare(
		r.Exchange,
		// 交换机类型，fanout广播类型
		"direct",
		true,
		false,
		// internal为true表示这个exchange不可以被client用来推送消息，仅用来进行
		// exchange和exchange之间的绑定
		false,
		false,
		nil,
	)

	r.failOnErr(err, "Failed to declare an exchange")

	// 2.试探性创建队列，这里注意队列名称不要写
	q, err := r.channel.QueueDeclare(
		"", // 随机生成队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	// 3.绑定队列到 exchange中
	err = r.channel.QueueBind(
		q.Name,
		// 在pub/sub模式下，这里的key要为空,即随机绑定交换机
		r.Key,
		r.Exchange,
		false,
		nil,
	)

	// 消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)

	go func() {
		for d := range messages {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Println("退出请按 CTRL+C\n")
	<-forever
}

// 路由模式
// 创建RabbitMQ实例
func NewRabbitMQRouting(exchangeName string, routingKey string) *RabbitMQ {
	// 创建RabbitMQ实例
	rabbitmq := NewRabbitMQ("", exchangeName, routingKey)
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建连接错误！")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败！")
	return rabbitmq
}

// 路由模式发送消息
