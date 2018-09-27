---
title: Tại sao dùng terraform
slug: tai-sao-dung-terraform
author: mark
published: 2018-08-09T11:57:00+07:00
tags:
  - devops
---

The originalarticle is Why we use Terraform and not Chef, Puppet, Ansible, SaltStack, or CloudFormation. Besides translating it to Vietnamese, I also performed some editing on the content and added some of my own notes to it.

You can also read the original article, at: https://blog.gruntwork.io/why-we-use-terraform-and-not-chef-puppet-ansible-saltstack-or-cloudformation-7989dad2865c if you have time to spare.

Đây là bài dịch từ Why we use Terraform and not Chef, Puppet, Ansible, SaltStack, or CloudFormation. Ngoài việc dịch sang tiếng Viết, mình còn sửa một số chỗ và thêm một ít ghi chú (của cá nhân mình) nữa. Những phần ghi chú sẽ được note lại như thế này:

> TN: Đây là một ghi chú của người dịch

Bạn có thể đọc bài viết gốc tại: https://blog.gruntwork.io/why-we-use-terraform-and-not-chef-puppet-ansible-saltstack-or-cloudformation-7989dad2865c

******

Hiện tại, nếu bạn tìm trên Google với từ khóa “infrastructure-as-code”, bạn sẽ thấy một số công cụ phổ biến như là:

- Chef
- Puppet
- Ansible
- SaltStack
- CloudFormation
- Terraform

Tất cả những phần mềm này đều có thể giúp bạn quản lý infrastructure (servers, virtual private network - VPC, storage, etc) trên nhiều cloud provider khác nhau (Google Cloud, AWS, Azure, etc), đều open source, được cập nhập thường xuyên và có documentation / tutorial khá đầy đủ. Vậy nếu bạn có nhu cầu quản lý infrastructure thì chọn cái nào?

Sau đây là bài phân tích của Yevgeniy Brikman, engineer tại Grunkwork.io về lý do họ chọn Hashicorp Terraform để quản lý infrastructure của công ty này:


Cách phân tích của Yevgeniy là so sánh những phần mềm trên thông qua 4 ý tưởng chính:
- Configuration management vs Orchestration
- Mutable Infrastructure vs Immutable Infrastructure
- Procedural vs Declarative
- Client/Server Architecture vs Client-only Architecture

## Configuration Management vs Orchestration

> TN:
> - ví dụ của configuration management là mỗi lần có patch của OpenSSL, etc, bạn sẽ update những phần mềm, package có liên quan server của mình (e.g: libcurl)
> - orchestration có thể tạm hiểu là bạn không tự quản lý server của mình, mà sẽ dùng 1 phần mềm khác để quản lý (e.g: Kubernetes là 1 dạng orchestration software, vì bạn không trực tiếp download docker images & chạy chúng).

Chef, Puppet, Ansible và SaltStack đều là configuration management tool. Chúng được thiết kế để cài đặt và quản lý phần mềm trên những server có sẵn. Ngược lại, CloudFormation và Terraform là orchestration tool, được thiết kế để tạo servers. Đây là hai công việc tương đối khác nhau, nhưng vẫn có những chỗ tương tự. Thông thường thì configuration management tool cũng có thể tạo server và orchestration tool cũng có thể cài đặt phần mềm một cách đơn giản. Tuy nhiên tùy vào việc cụ thể bạn cần, một số phần mềm sẽ dễ dùng hơn rất nhiều.

Ví dụ, nếu bạn deploy containers (e.g: dùng Docker), hầu như bạn sẽ không cần phải manage configuration. Trong docker image đã có những gì bạn cần để chạy chương trình của mình. Vì vậy, bạn chỉ cần provision servers để chạy những container này (và chạy docker daemon trên server). Orchestration tool như Terraform thường sẽ làm được việc này dễ dàng hơn.

## Mutable Infrastructure vs Immutable Infrastructure

> TN:
> - mutable: có thể thay đổi; immutable: không thay đổi được. Có thể hiểu đại khái giống như sự khác nhau giữa variable && constant trong hầu hết các ngôn ngữ lập trình

Configuration management tool như Chef, Puppet, Ansible và SaltStack thường hoạt động bằng cách thay đổi hệ thống có sẵn. Ví dụ như nếu bạn muốn cài version mới của OpenSSL trên một server sẵn có, những tool này sẽ update OpenSSL trực tiếp trên server của bạn (e.g: `apt-get upgrade`). Khi bạn apply nhiều update hơn, mỗi server có thể có những trạng thái tồn tại khác nhau. Điều này sẽ làm cho vấn đề “configuration drift” (mỗi server sẽ hơi khác với cấu hình chuẩn mà bạn muốn) xuất hiện thường xuyên hơn và việc này thường rất khó để phát hiện và xử lý.

> TN: vấn đề configuration drift thường xuất hiện rõ ràng hơn khi bạn có >= 2 người thường xuyên update server song song (concurrently).

Ngược lại, nếu bạn dùng orchestration tool như Terraform để deploy server, mỗi thay đổi sẽ tạo ra một deployment mới. Cũng dùng ví dụ deploy OpenSSL như ở trên, bạn thường sẽ tạo ra 1 machine image khác (e.g bằng Hashicorp Packer) và dùng bản OpenSSL mới này. Sau đó bạn sẽ deploy server với image mới và xóa server cũ. Phương pháp này giảm khả năng configuration drift xuất hiện và sẽ làm bạn sẽ biết chắc trạng thái hiện tại của servers hơn.

> TN: không phải lúc nào cũng có thể dễ dàng deploy server mới && xóa server cũ, ví dụ như nếu đó là DB server.


## Procedural vs Declarative

> TN: hãy nghĩ về sự khác nhau của
>
>     sum := 0;
>     for value in myList:
>         sum += value
>     # procedural
>    
>     và
>    
>     sum := sum(myList, 0)
>     # declarative

Với Chef và Ansible bạn thường phải viết procedural code cho từng thay đổi trên server, sau đó chạy theo 1 thứ tự nhất định để đạt được cấu hình mà bạn muốn. Ngược lại, dùng Terraform, CloudFormation, SaltStack và Puppet thì bạn cần declare cấu hình cuối cùng, sau đó chúng sẽ tự tìm cách để setup server như vậy.

Ví dụ cụ thể hơn:

- Bạn muốn chạy 10 servers (EC2 instances trên AWS) và chạy v1 của application. Bạn sẽ dùng Ansible template như sau:

```
- ec2:
    count: 10
    image: ami-v1
    instance_type: t2.micro
```


Còn với Terraform, bạn viết làm như thế này:


```
resource "aws_instance" "example" {
  count = 10
  ami = "ami-v1"
  instance_type = "t2.micro"
}
```

Phần này thì tương đối giống nhau.

Tiếp tục, nếu bạn cần chuyển từ 10 lên 15 server. Với Ansible, bạn phải biết là hiện tại có 10 server rồi, và cần deploy thêm 5 cái nữa, rồi viết:


```
- ec2:
    count: 5
    image: ami-v1
    instance_type: t2.micro
```

Ngược lại, với Terraform, bạn chỉ cần sửa lại:


```
resource "aws_instance" "example" {
  count = 15
  ami = "ami-v1"
  instance_type = "t2.micro"
}
```

Terraform sẽ kiểm tra hệ thống hiện tại của bạn và biết là cần thêm 5 server nữa. Bạn cũng có thể dry-run để kiểm tra xem Terraform sẽ làm gi:


```
$> terraform plan
+ aws_instance.example.11
    ami:                      "ami-v1"
    instance_type:            "t2.micro"
+ aws_instance.example.12
    ami:                      "ami-v1"
    instance_type:            "t2.micro"
+ aws_instance.example.13
    ami:                      "ami-v1"
    instance_type:            "t2.micro"
+ aws_instance.example.14
    ami:                      "ami-v1"
    instance_type:            "t2.micro"
+ aws_instance.example.15
    ami:                      "ami-v1"
    instance_type:            "t2.micro"
Plan: 5 to add, 0 to change, 0 to destroy.
```

Tiếp tục, nếu bạn muốn chạy v2 của app trên những server này thì sao?

Với Ansible, bạn sẽ phải tìm những server mà bạn đã deployed (10 hay 15 nhỉ?), sau đó update mỗi server.

Còn với Terraform, bạn sẽ viết:


```
resource "aws_instance" "example" {
  count = 15
  ami = "ami-v2"
  instance_type = "t2.micro"
}
```

Đây là một ví dụ khá đơn giản, và trên thực tế thì cũng có cách làm Ansible hoạt động giống Terraform (thông qua `instance_tags` và `count_tag`). Tuy nhiên việc này sẽ phức tạp hơn rất nhiều.

Tóm lại, sự khác biệt chính giữa procedural và declarative tool là:

Procedural tool thường chỉ lưu sự thay đổi (changes), bạn phải biết những thay đổi này được apply theo thứ tự nào thì mới biết trạng thái hiện tại của hệ thống. Điều này làm cho code của procedural tool khó sử dụng lại hơn và trở nên phức tạp trên những hệ thống lớn. Lý do chính là của việc này là bạn phải biết trạng thái hiện tại của hệ thống để hiểu được code sẽ làm gì.

Điểm mạnh của declarative tool là code luôn thể hiện trạng thái hiện tại của hệ thống. Bạn chỉ cần tập trung vào việc miêu tả trạng thái mình mong muốn thôi.


Tuy nhiên declarative tool cũng có những điểm yếu riêng của nó. Bởi vì không dùng programming language, bạn chỉ có thể configure hệ thống theo những cách nhất định. Thêm vào đó, một số việc sẽ rất khó để viết declarative code, ví dụ như rolling update, zero-downtime deployment chẳng hạn. Về mặt này thì Terraform có một số chức năng như input variables, modules, create_before_destroy, count và interpolation functions, để giúp việc configuration dễ dàng hơn.

## Client/Server Architecture vs Client-Only Architecture

Chef, Puppet và SaltStack dùng client/server architecture by default. Bạn sẽ dùng client (web UI / CLI tool) để viết command (e.g: `deploy X`). Command này sẽ được server nhận, và server sẽ chạy và track những thay đổi của hệ thống. Để chạy command, server sẽ gọi những chương trình agent chạy trên từng server. Thiết kế này có một số vấn đề chính:

- Bạn phải cài thêm chương trình (agent) trên mỗi server và sẽ phải monitor, upgrade những chương trình này
- Bạn phải chạy thêm 1 (hoặc nhiều server, nếu cần high availability) để manage configuration
- Để server và agent có thể communicate, bạn có thể phải mở thêm port và nhiều thứ khác, làm tăng attack surface của hệ thống.


Vì những lý do trên, hệ thống của bạn sẽ có thể fail theo nhiều cách khác nữa (vd: không deploy được / server xác nhận deploy nhưng agent không chạy command, etc).

CloudFormation, Ansible và Terraform dùng architecture thiên về client-only, nên có thể ít gặp những vấn đề trên hơn.


> TN: CloudFormation cũng dùng client/server architecture, nhưng vì server được chạy bời AWS nên người dùng thường chỉ cần lo về client code thôi. Tương tự, Terraform có khái niệm remote state, chính là phần "server", nhưng nó cũng có thể chỉ chạy trên local.
Ansible client dùng ssh để connect thẳng vào server còn Terraform dùng API của cloud provider để provision infrastructure, vì vậy mà dùng Terraform, thường bạn không cần cài thêm quá nhiều thứ.

## Kết luận

Dưới đây là so sánh tổng quát của một số IAC tool:

|                | Chef          | Puppet        | Ansible     | SaltStack    | CloudFormation | Terraform     |
|----------------|---------------|---------------|-------------|--------------|----------------|---------------|
| Code           | Open source   | Open source   | Open source | Open source  | Closed source  | Open source   |
| Clould         | All           | All           | All         | All          | AWS only       | All           |
| Type           | Config Mgmt   | Config Mgmt   | Config Mgmt | Config Mgmt  | Orchestration  | Orchestration |
| Infrastructure | Mutable       | Mutable       | Mutable     | Mutable      | Immutable      | Immutable     |
| Language       | Procedural    | Declarative   | Procedural  | Declarative  | Declarative    | Declarative   |
| Architecture   | Client/Server | Client/Server | Client-Only | Client/Serve | Client-Only    | Client-Only   |

Vì Grunkwork muốn tìm một open source, cloud-agnostic orchestration tool, support for immutable infrastructure, dùng declarative language và client-only architecture nên team này đã chọn Terraform.

So với Puppet (từ 2005), Chef (từ 2009), SaltStack, CloudFormation (từ 2011) và Ansible (từ 2012), Terraform (từ 2014) vẫn còn khá mới. Nó vẫn còn nhiều bug và có vấn đề với việc chia sẻ trạng thái hệ thống (remote state management). Tuy vậy có thể những điểm mạnh của nó đủ để bù đắp cho những điểm yếu này.

> TN: Bạn có thể tìm hiểu thêm về Terraform theo qua:
> - Software Engineer Radio: Terraform and Declarative Programming
>    http://www.se-radio.net/2017/04/se-radio-episode-289-james-turnbull-on-declarative-programming-with-terraform/
> - The Terraform book:
>    https://terraformbook.com/
