module github.com/gmcorenet/bundle-crud

go 1.26.2

require (
	github.com/gmcorenet/sdk/gmcore-bundle v0.0.0
	github.com/gmcorenet/sdk/gmcore-crud v0.0.0
)

require (
	github.com/gmcorenet/sdk/gmcore-form v0.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/text v0.2.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/gorm v1.25.10 // indirect
)

replace (
	github.com/gmcorenet/sdk/gmcore-bundle => ../../sdks/gmcore-bundle
	github.com/gmcorenet/sdk/gmcore-crud => ../../sdks/gmcore-crud
	github.com/gmcorenet/sdk/gmcore-form => ../../sdks/gmcore-form
	github.com/gmcorenet/sdk/gmcore-i18n => ../../sdks/gmcore-i18n
	github.com/gmcorenet/sdk/gmcore-router => ../../sdks/gmcore-router
	github.com/gmcorenet/sdk/gmcore-security => ../../sdks/gmcore-security
	github.com/gmcorenet/sdk/gmcore-templating => ../../sdks/gmcore-templating
)
