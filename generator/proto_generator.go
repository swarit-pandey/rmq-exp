package generator

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Generator is random mumbo jumbo generator for proto definitions/types
type Generator struct {
	randSeed     *rand.Rand
	config       GeneratorConfig
	messageTypes map[string]protoreflect.MessageDescriptor
}

type GeneratorConfig struct {
	StringLength    int
	BytesLength     int
	MaxListElements int
	MaxMapElements  int
}

func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		StringLength:    10,
		BytesLength:     20,
		MaxListElements: 5,
		MaxMapElements:  5,
	}
}

// NewGenerator instantiates a new proto message generator, pass path to
// proto file here
func NewGenerator(protoPath string) (*Generator, error) {
	if _, err := os.Stat(protoPath); err != nil {
		return nil, fmt.Errorf("proto file not found: %v", err)
	}

	gen := &Generator{
		randSeed:     rand.New(rand.NewSource(time.Now().UnixNano())),
		config:       DefaultConfig(),
		messageTypes: make(map[string]protoreflect.MessageDescriptor),
	}

	protoregistry.GlobalTypes.RangeMessages(func(mt protoreflect.MessageType) bool {
		desc := mt.Descriptor()
		gen.messageTypes[string(desc.FullName())] = desc
		return true
	})

	if len(gen.messageTypes) == 0 {
		return nil, fmt.Errorf("no message types found in proto files")
	}

	return gen, nil
}

func (g *Generator) generateRandomValue(fd protoreflect.FieldDescriptor) any {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return g.randSeed.Intn(2) == 1
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int32(g.randSeed.Int31())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return int64(g.randSeed.Int63())
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32(g.randSeed.Uint32())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return uint64(g.randSeed.Uint64())
	case protoreflect.FloatKind:
		return float32(g.randSeed.Float64())
	case protoreflect.DoubleKind:
		return g.randSeed.Float64()
	case protoreflect.StringKind:
		return g.randomString(10)
	case protoreflect.BytesKind:
		return g.randomBytes(20)
	case protoreflect.EnumKind:
		values := fd.Enum().Values()
		return values.Get(g.randSeed.Intn(values.Len())).Number()
	case protoreflect.MessageKind:
		return g.randomMessage(fd.Message())
	default:
		return nil
	}
}

func (g *Generator) randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[g.randSeed.Intn(len(charset))]
	}
	return string(b)
}

func (g *Generator) randomBytes(length int) []byte {
	b := make([]byte, length)
	g.randSeed.Read(b)
	return b
}

// TODO: Way too fucking hardcoded for summary-v2 proto, make it generic

func (g *Generator) randomMessage(md protoreflect.MessageDescriptor) proto.Message {
	msg := dynamicpb.NewMessage(md)
	fields := md.Fields()

	if md.FullName() == "summary.SummaryEvent" {
		metadataField := fields.ByName("metadata")
		metadata := g.randomMessage(metadataField.Message())
		metadataMsg := metadata.ProtoReflect()
		metadataFields := metadataField.Message().Fields()
		metadataMsg.Set(metadataFields.ByName("cluster_id"), protoreflect.ValueOf(int64(12345)))
		metadataMsg.Set(metadataFields.ByName("tenant_id"), protoreflect.ValueOf(int64(67890)))
		msg.Set(metadataField, protoreflect.ValueOf(metadata))

		msg.Set(fields.ByName("namespace"), protoreflect.ValueOf("artificial-namespace"))
		msg.Set(fields.ByName("operation"), protoreflect.ValueOf("artificial-operation"))
		msg.Set(fields.ByName("action"), protoreflect.ValueOf("artificial-action"))

		workloadField := fields.ByName("workload")
		workload := dynamicpb.NewMessage(workloadField.Message())
		workloadFields := workloadField.Message().Fields()
		workload.Set(workloadFields.ByName("type"), protoreflect.ValueOf("artificial-workload-type"))
		workload.Set(workloadFields.ByName("name"), protoreflect.ValueOf("artificial-workload-name"))
		msg.Set(workloadField, protoreflect.ValueOf(workload))

		msg.Set(fields.ByName("operation"), protoreflect.ValueOf("Process"))
		msg.Set(fields.ByName("action"), protoreflect.ValueOf("artificial-action"))
		msg.Set(fields.ByName("namespace"), protoreflect.ValueOf("artificial-namespace"))

		processFileField := fields.ByName("process_file")
		if processFileField != nil {
			processFile := dynamicpb.NewMessage(processFileField.Message())
			processFileFields := processFileField.Message().Fields()

			processFile.Set(processFileFields.ByName("pod"), protoreflect.ValueOf("artificial-pod"))
			processFile.Set(processFileFields.ByName("source"), protoreflect.ValueOf("artificial-source"))
			processFile.Set(processFileFields.ByName("destination"), protoreflect.ValueOf("artificial-destination"))
			processFile.Set(processFileFields.ByName("count"), protoreflect.ValueOf(int64(1)))
			processFile.Set(processFileFields.ByName("updated_time"), protoreflect.ValueOf(time.Now().Unix()))
			processFile.Set(processFileFields.ByName("action"), protoreflect.ValueOf("artificial-action"))

			containerField := processFileFields.ByName("container")
			if containerField != nil {
				container := dynamicpb.NewMessage(containerField.Message())
				containerFields := containerField.Message().Fields()
				container.Set(containerFields.ByName("name"), protoreflect.ValueOf("artificial-container"))
				container.Set(containerFields.ByName("image"), protoreflect.ValueOf("nginx:latest"))
				processFile.Set(containerField, protoreflect.ValueOf(container))
			}

			msg.Set(processFileField, protoreflect.ValueOf(processFile))
		}
		return msg
	}

	switch md.FullName() {
	case "summary.SummaryEventMetadata":
		msg.Set(fields.ByName("cluster_id"), protoreflect.ValueOf(int64(g.randSeed.Int63n(1000000))))
		msg.Set(fields.ByName("tenant_id"), protoreflect.ValueOf(int64(g.randSeed.Int63n(1000000))))

	case "summary.SummaryEventWorkload":
		msg.Set(fields.ByName("type"), protoreflect.ValueOf(g.randomString(10)))
		msg.Set(fields.ByName("name"), protoreflect.ValueOf(g.randomString(10)))
	}

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.IsList() {
			list := msg.Mutable(fd).List()
			numElements := g.randSeed.Intn(5) + 1
			for j := 0; j < numElements; j++ {
				value := g.generateRandomValue(fd)
				list.Append(protoreflect.ValueOf(value))
			}
		} else if fd.IsMap() {
			mp := msg.Mutable(fd).Map()
			numElements := g.randSeed.Intn(5) + 1
			for j := 0; j < numElements; j++ {
				key := g.generateRandomValue(fd.MapKey())
				value := g.generateRandomValue(fd.MapValue())
				mp.Set(protoreflect.MapKey(protoreflect.ValueOf(key)), protoreflect.ValueOf(value))
			}
		} else {
			value := g.generateRandomValue(fd)
			msg.Set(fd, protoreflect.ValueOf(value))
		}
	}

	return msg
}

func (g *Generator) generateProcessFileEvent(md protoreflect.MessageDescriptor) proto.Message {
	msg := dynamicpb.NewMessage(md)
	fields := md.Fields()

	msg.Set(fields.ByName("pod"), protoreflect.ValueOf(g.randomString(10)))
	msg.Set(fields.ByName("source"), protoreflect.ValueOf(g.randomString(10)))
	msg.Set(fields.ByName("destination"), protoreflect.ValueOf(g.randomString(10)))
	msg.Set(fields.ByName("count"), protoreflect.ValueOf(int64(g.randSeed.Int63n(100))))
	msg.Set(fields.ByName("updated_time"), protoreflect.ValueOf(time.Now().Unix()))
	msg.Set(fields.ByName("action"), protoreflect.ValueOf(g.randomString(10)))

	containerField := fields.ByName("container")
	if containerField != nil {
		container := g.generateContainer(containerField.Message())
		msg.Set(containerField, protoreflect.ValueOf(container))
	}

	return msg
}

func (g *Generator) generateContainer(md protoreflect.MessageDescriptor) proto.Message {
	msg := dynamicpb.NewMessage(md)
	fields := md.Fields()

	msg.Set(fields.ByName("name"), protoreflect.ValueOf(g.randomString(10)))
	msg.Set(fields.ByName("image"), protoreflect.ValueOf("nginx:latest"))

	return msg
}

func (g *Generator) ensureWorkloadFields(md protoreflect.MessageDescriptor) proto.Message {
	msg := dynamicpb.NewMessage(md)
	fields := md.Fields()

	msg.Set(fields.ByName("type"), protoreflect.ValueOf(g.randomString(10)))
	msg.Set(fields.ByName("name"), protoreflect.ValueOf(g.randomString(10)))
	msg.Set(fields.ByName("labels"), protoreflect.ValueOf(g.randomString(10)))

	return msg
}

// ListMessageTypes gets you all the message types available
func (g *Generator) ListMessageTypes() []string {
	types := make([]string, 0, len(g.messageTypes))
	for tn := range g.messageTypes {
		types = append(types, tn)
	}
	return types
}

// GenerateMessage will generate random values for proto fields
func (g *Generator) GenerateMessage(msgType string) (protoreflect.ProtoMessage, error) {
	desc, ok := g.messageTypes[msgType]
	if !ok {
		return nil, fmt.Errorf("message type %s not found", msgType)
	}
	return g.randomMessage(desc), nil
}
