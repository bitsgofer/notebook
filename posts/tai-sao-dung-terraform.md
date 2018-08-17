---
title: Tại sao dùng terraform
slug: tai-sao-dung-terraform
author: mark
published: 2018-08-09T11:57:00+07:00
tags:
  - devops
---

# NOTE: this article is still a working-in-progress

> The orignal article is [Why we use Terraform and not Chef, Puppet, Ansible, SaltStack, or CloudFormation](https://blog.gruntwork.io/why-we-use-terraform-and-not-chef-puppet-ansible-saltstack-or-cloudformation-7989dad2865c).
I only translated it to Vietnamese and performed some editing on the content.
>
> Đây là bài dịch từ [Why we use Terraform and not Chef, Puppet, Ansible, SaltStack, or CloudFormation](https://blog.gruntwork.io/why-we-use-terraform-and-not-chef-puppet-ansible-saltstack-or-cloudformation-7989dad2865c).
Ngoài việc dịch sang tiếng Viết, mình còn sửa 1 số chỗ cho dễ đọc hơn.

Hiện tại, nếu bạn tìm công cụ để quản lý infrastructure với từ khóa "infrastructure-as-code", bạn sẽ thấy một số công cụ phổ biến như là:

- Chef
- Puppet
- Ansible
- SaltStack
- CloudFormation
- Terraform

Tất cả những tool ở trên đều có thể giúp bạn quản lý infrastructure trên các cloud provider khác nhau, đều open source, được cập nhập thường xuyên và có documentation / tutorial khá đầy đủ. Vậy thì nếu bạn cần chọn tool để dùng thì chọn cái nào?

Sau đây là bài phân tích của engineer @ Grunkwork.io về lý do họ chọn Terraform (của Hashicorp):

## Configuration Management vs Orchestration

Chef, Puppet, Ansible và SaltStack đều là "configuration management tool". Chúng được thiết kế để install và manage software trên những server có sãn. Ngược lại, CloudFormation và Terraform là "orchestration tool", được thiết kế để provision servers. Đây là hai công việc tương đối khác nhau, nhưng vẫn có những chỗ tương tự. Thông thường thì "configuration management tool" cũng có thể provision servers và orchestration tool cũng có thể provide configuration management to some degree. Vì vậy mà bạn cần chọn tool tùy theo việc bạn muốn làm.

Cụ thể là, nếu bạn deploy containers (e.g: Docker), hầu như bạn sẽ không cần phải manage configuration. Trong docker image sẽ có những gì bạn cần để chạy chương trình của mình. Vì vậy, bạn chỉ cần provision servers để chạy nhưng container này. Mà nếu như vậy thì một orchestration tool như Terraform thương sẽ làm được việc này.

// TODO: define: configuration management and orchestration

## Mutable Infrastructure vs Immutable Infrastructure

Configuration management tool như Chef, Puppet, Ansible và SaltStack thường hoạt động bằng cách thay đổi hẹ thống có sẵn. Ví dụ như nếu bạn muốn cài version mới của OpenSSL, những tool này sẽ update OpenSSL trên server của bạn. Cùng với việc bạn apply nhiều update hơn, mỗi server sẽ có những trạng thái tồn tại khác nhau. Điều này sẽ làm cho vấn đề "configuration drift" (mổi server sẽ hơi khác với cấu hình chuẩn mà bạn muốn) xuất hiện thường xuyên hơn và điều này thường rất khó để tìm và xử lý.

Ngược lại, nếu bạn dùng orchestration tool như Terraform để deploy server, mỗi thay đổi sẽ tạo ra một deployment mới. Ví dụ như cũng deploy new OpenSSL như ở trên, bạn sẽ tạo ra 1 image mới với bản OpenSSL này, sau đó thì sẽ deploy server mới và xóa server cũ. Phương pháp này giảm khả năng configuration drift xuất hiện và sẽ làm bạn dễ dàng biết trạng thái hiện tại của servers của mình.

// TODO: note: configuration management tool can also do this. Plus not all changes done by Terraform are immutable.

## Procedural vs Declarative

Với Chef và Ansible bạn thường phải viết code cho từng thay đổi trên server, để đạt được cấu hình mà bạn muốn. Terraform, CloudFormation, SaltStack và Puppet thì encourage bạn specify cấu hình cuối cùng, sau đó chúng sẽ tự tìm cách để setup server như vậy.

Ví dụ nếu bạn muốn deploy 10 servers (EC2 instances trên AWS) và chạy v1 của 1 app. Bạn sẽ làm điều đó với Ansible như sau:

```
- ec2:
    count: 10
    image: ami-v1
    instance_type: t2.micro
```

Còn với Terraform, bạn sẽ làm như thế này:

```
resource "aws_instance" "example" {
  count = 10
  ami = "ami-v1"
  instance_type = "t2.micro"
}
```

Phẩn này thì tương đối giốn nhau, nhưng đển lúc bạn cần thay đổi thì sẽ khác. Giả sử bạn cần chạy 15 server thay vì 10 như lúc đầu. Với Ansible, bạn phải biết là hiện tại có 10 server rồi, và cần deploy thêm 5 cái nữa:

```
- ec2:
    count: 5
    image: ami-v1
    instance_type: t2.micro
```

Ngược lại, với Terraform, bạn chỉ cần dùng:

```
resource "aws_instance" "example" {
  count = 15
  ami = "ami-v1"
  instance_type = "t2.micro"
}
```

Và Terraform sẽ kiểm tra hệ thống hiện tại của bạn và biết là cần thêm 5 server nữa. Bạn cũng có thể dry-run để kiểm tra xem Terraform sẽ làm gi:


<pre class="language-bash"><code class="language-bash">
> terraform plan
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
</code></pre>

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

Ansible cũng có thể được dùng để thực hiện công việc trên (thông qua `instance_tags` và `count_tag`). Tuy nhiên việc này sẽ phức tạp hơn rất nhiều.

Tóm lại, sự khác biệt chính giữa procedural và declarative tool là:

- Procedural tool thường chỉ lưu sụ thay đổi (changes), bạn phải biết những thay đổi này được apply theo thứ tự nào thì mới biết trạng thái hiện tại của hệ thống. Điều này làm cho code của procedural tool khó sử dụng lại hơn, vì để hiểu được code sẽ làm gì, bạn phải biết trạng thái hiện tại).
- Điểm mạnh của declarative tool là code luôn thể hiện trạng thái hiện tại của hệ thống. Bạn chỉ cần tập trung vào việc miêu tả trạng thái mình mong muốn thôi.

Tuy nhiên declarative tool cũng có những điểm yếu riêng của nó. Bởi vì không dùng programming language, bạn chỉ có thể configure hệ thống theo những cách nhất định. Thêm vào đó, một sổ task sẽ rất khó để mô tả declaratively, ví dụ như rolling update, zero-downtime deployment chẳng hạn. Về mặt này thì Terraform có một số chức năng như input variables, modules, create_before_destroy, count và interpolation functions, để giúp việc configuration dễ dàng hơn.


## Client/Server Architecture vs Client-Only Architecture

Chef, Puppet và SaltStack dùng client/server architecture by default. Bạn sẽ dùng client (web UI / CLI tool) để issue command (e.g: "deploy X"). Command này sẽ được server nhận, chạy và server sẽ keep track of the state of the system. Để chạy command, server sẽ talk to những chương trình agent chạy trên từng server. Approach này có một số vấn đề:

- Bạn phải cài thêm chương trình (agent) trên mỗi server, và sẽ phải monitor, upgrade những chương trình này
- Bạn phải chạy thêm 1 hoặc nhiều server để manage configuration
- Để server và agent có thể communicate, bạn có thể phải mở thêm port và nhiêu thứ khác, làm tăng attack surface của hệ thống.

Vì có thêm nhiều thứ, hệ thống của bạn sẽ có thể fail theo nhiều cách khác nữa (vd: không deploy được / deploy nhưng agent không chạy command, etc).

CloudFormation, Ansible và Terraform dùng architecture giống như client-only, nên có thể không gặp những vấn đề trên:

- CloudFormation cũng dùng client/server architecture, nhưng bì AWS handle server nên ngươi dùng chỉ cần lo về client code thôi.
- Ansible client dùng ssh để connect thẳng vào server.
- Terraform dùng API của cloud provider để provision infrastructure, vì vậy mà không có thêm quá nhiều thứ. Đây có vẻ là option tốt nhất về eas-of-use, security và maintainability.

Kết luận:

Dưới đây là so sánh tổng quát của một số IAC tool:

|                | Chef          | Puppet        | Ansible     | SaltStack    | CloudFormation | Terraform     |
|----------------|---------------|---------------|-------------|--------------|----------------|---------------|
| Code           | Open source   | Open source   | Open source | Open source  | Closed source  | Open source   |
| Clould         | All           | All           | All         | All          | AWS only       | All           |
| Type           | Config Mgmt   | Config Mgmt   | Config Mgmt | Config Mgmt  | Orchestration  | Orchestration |
| Infrastructure | Mutable       | Mutable       | Mutable     | Mutable      | Immutable      | Immutable     |
| Language       | Procedural    | Declarative   | Procedural  | Declarative  | Declarative    | Declarative   |
| Architecture   | Client/Server | Client/Server | Client-Only | Client/Serve | Client-Only    | Client-Only   |

Vì Grunkwork muốn tìm một open source, cloud-agnostic orchestration tool với support for immutable infrastructure,
dùng declarative language và client-only architecture nên team này đã chọn Terraform.

Terraform isn't perfect. So với Puppet có từ 2005, Chef từ 2009 SaltStack, CloudFormation từ 201 và Ansible từ 2012,
Terraform khá mới (từ 2014), vần còn nhiều bug và có vấn đề với việc storing state. Tuy vậy nó nhưng điểm mạnh của nó
có thể bù đắp cho những điểm yếu.
