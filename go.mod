module github.com/yubing744/trading-gpt

go 1.20

require (
	github.com/c9s/bbgo v1.43.1
	github.com/dop251/goja v0.0.0-20230304130813-e2f543bf4b4c
	github.com/google/uuid v1.3.0
	github.com/kataras/go-events v0.0.3
	github.com/larksuite/oapi-sdk-go/v3 v3.0.17
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.2
	github.com/tmc/langchaingo v0.0.0-20230429215204-931a1fe7a2a9
)

replace github.com/c9s/bbgo => ./libs/bbgo
replace github.com/tmc/langchaingo => ./libs/langchain

require (
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/VividCortex/ewma v1.1.1 // indirect
	github.com/adshao/go-binance/v2 v2.3.10 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/c-bata/goptuna v0.8.1 // indirect
	github.com/c9s/requestgen v1.3.0 // indirect
	github.com/c9s/rockhopper v1.2.2-0.20220617053729-ffdc87df194b // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cheggaaa/pb/v3 v3.0.8 // indirect
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/codingconcepts/env v0.0.0-20200821220118-a8fbf8d84482 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisenkom/go-mssqldb v0.12.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gertd/go-pluralize v0.2.1 // indirect
	github.com/gin-contrib/cors v1.3.1 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.7.0 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-redis/redis/v8 v8.8.0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/heroku/rollrus v0.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jedib0t/go-pretty/v6 v6.3.6 // indirect
	github.com/jmoiron/sqlx v1.3.4 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leekchan/accounting v0.0.0-20191218023648-17a4ce5f94d4 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.2.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v1.14.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/clusters v0.0.0-20180605185049-a07a36e67d36 // indirect
	github.com/muesli/kmeans v0.3.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/otp v1.3.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.0 // indirect
	github.com/rollbar/rollbar-go v1.4.5 // indirect
	github.com/sajari/regression v1.0.1 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/slack-go/slack v0.10.1 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	
	github.com/ugorji/go/codec v1.2.3 // indirect
	github.com/valyala/fastjson v1.5.1 // indirect
	github.com/wcharczuk/go-chart/v2 v2.1.0 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.opentelemetry.io/otel v0.19.0 // indirect
	go.opentelemetry.io/otel/metric v0.19.0 // indirect
	go.opentelemetry.io/otel/trace v0.19.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/image v0.0.0-20200927104501-e162460cd6b5 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	golang.org/x/tools v0.1.12 // indirect
	gonum.org/v1/gonum v0.8.2 // indirect
	google.golang.org/genproto v0.0.0-20220405205423-9d709892a2bf // indirect
	google.golang.org/grpc v1.45.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/tucnak/telebot.v2 v2.5.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
