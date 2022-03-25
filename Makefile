gen_byteair:
	protoc --go_out=byteair/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_byteair.proto

gen_common:
	protoc --go_out=common/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_common.proto

gen_general:
	protoc --go_out=general/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_general.proto

gen_retail:
	protoc --go_out=retail/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_retail.proto

gen_retail2:
	protoc --go_out=retailv2/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_retailv2.proto

gen_media:
	protoc --go_out=media/protocol -I=docs --go_opt=paths=source_relative docs/byteplus_media.proto