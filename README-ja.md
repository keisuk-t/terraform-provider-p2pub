# terraform-provider-p2pub

Terraform Custom Provider for P2 PUB

# 使い方

## セットアップ

### ビルドとインストール

make してバイナリを作成した後、.terraformrc にパスを追加します

```
$ make build
$ cp terraform-provider-p2pub /path/to/install/dir
```

- ~/.terraformrc

```
providers {
    p2pub = "/path/to/install/dir/terraform-provider-p2pub"
}
```

### プロバイダーの設定

tf ファイルに API キーと gis サービスコードを記述します。Terraform で管理されるリソース(仮想サーバー等)はここで指定した gis サービスコード配下の契約になります。

```
provider "p2pub" {
    access_key_id = "<ACCESSKEY>"
    secret_access_key = "<SECRETKEY>"
    gis_service_code = "<SERVICECODE>"
}
```

設定値は環境変数から与えることもできます。

|項目|環境変数名|
|-|-|
|```access_key_id```|```IIJAPI_ACCESS_KEY```|
|```secret_access_key```|```IIJAPI_SECRET_KEY```|
|```gis_service_code```|```GISSERVICECODE```|

## Terraform 実行

固有の手順は特にありません。通常通り ```terraform plan```, ```terraform apply```, ```terraform destroy```, etc... を実行してください。

設定例は ```example/``` を参照してください。

## リソース一覧

### ```p2pub_virtual_server```

[仮想サーバー](http://manual.iij.jp/p2/pub/b-1.html)

|項目|内容|値|必須|
|-|-|-|-|
|```type```|仮想サーバー品目| VBxx-xx, VGxx-xx, VDxx-xx |◯|
|```os_type```|OS種別|Linux, Windows|◯|
|```label```|ラベル|任意の文字列||
|```system_storage```|接続するブートデバイス|ibaサービスコード||
|```data_storage```|接続する追加ストレージ|ibb/ibgサービスコードのリスト||
|```private_network```|接続するプライベートネットワーク/V|ivlサービスコードのリスト||
|```enable_global_ip```|グローバルIPアドレス使用の有無|true/false||


```
resource "p2pub_virtual_server" "vs1" {
    type = "VB0-1"
    os_type = "Linux"
    label = "vs-label"
    system_storage = "iba########"
    data_storage = ["ibg#######"]
    private_network = ["ivl########"]
    enable_global_ip = true
}
```

### ```p2pub_system_storage```

[システムストレージ](http://manual.iij.jp/p2/pub/b-3-1.html)

|項目|内容|値|必須|
|-|-|-|-|
|```type```|システムストレージ品目| http://manual.iij.jp/p2/pubapi/59949023.html |◯|
|```label```|ラベル|任意の文字列||
|```root_ssh_key```|rootのSSH公開鍵|||
|```root_password```|rootパスワード|||

```
resource "p2pub_system_storage" "ss1" {
    type = "S30GB_CENTOS7_64"
    label = "ss-label"
    root_ssh_key = "<SSH_PUBLIC_KEY>"
}
```

### ```p2pub_additional_storage```

[追加ストレージ](http://manual.iij.jp/p2/pub/b-3-1.html)

|項目|内容|値|必須|
|-|-|-|-|
|```type```|システムストレージ品目| http://manual.iij.jp/p2/pubapi/59949023.html |◯|
|```label```|ラベル|任意の文字列||

```
resource "p2pub_additional_storage" "as1" {
    type = "B1000GB"
    label = "as-label"
}
```

### ```p2pub_private_network```

[プライベートネットワーク/V](http://manual.iij.jp/p2/pub/b-5-1-1.html)

|項目|内容|値|必須|
|-|-|-|-|
|```label```|ラベル|任意の文字列||

```
resource "p2pub_private_network" "net1" {
    label = "pn-label"
}
```

### ```p2pub_storage_archive```

[ストレージアーカイブ](http://manual.iij.jp/p2/pub/b-4.html)

(設定項目なし)

```
resource "p2pub_storage_archive" "sa1" {}
```


### ```p2pub_global_ip_address```

[グローバルIPアドレス](http://manual.iij.jp/p2/pub/b-5-1-2.html)

|項目|内容|値|必須|
|-|-|-|-|
|```address_num```|IPアドレスの契約数|1~20までの整数|◯|

```
resource "p2pub_global_ip_address" "ip1" {
    address_num = 10
}
```
