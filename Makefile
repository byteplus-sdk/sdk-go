gen_common:
	protoc --go_out=common/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_common.proto

gen_general:
	protoc --go_out=general/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_general.proto

gen_retail:
	protoc --go_out=retail/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_retail.proto