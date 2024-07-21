<p align="center"><img width="450" src="./xgen.svg" alt="xgen logo"></p>

<br>

<p align="center">
    <a href="https://github.com/xuri/xgen/actions?workflow=Go"><img src="https://github.com/xuri/xgen/workflows/Go/badge.svg?branch=master" alt="Build Status"></a>
    <a href="https://codecov.io/gh/xuri/xgen"><img src="https://codecov.io/gh/xuri/xgen/branch/master/graph/badge.svg" alt="Code Coverage"></a>
    <a href="https://goreportcard.com/report/github.com/xuri/xgen"><img src="https://goreportcard.com/badge/github.com/xuri/xgen" alt="Go Report Card"></a>
    <a href="https://pkg.go.dev/github.com/xuri/xgen"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="go.dev"></a>
    <a href="https://opensource.org/licenses/BSD-3-Clause"><img src="https://img.shields.io/badge/license-bsd-orange.svg" alt="Licenses"></a>
    <a href="https://www.paypal.me/xuri"><img src="https://img.shields.io/badge/Donate-PayPal-green.svg" alt="Donate"></a>
</p>

# xgen

## Introduction

xgen 是 Go 语言编写的 XSD (XML Schema Definition) 工具基础库。使用本基础库要求使用的 Go 语言为 1.18 或更高版本，完整的 API 使用文档请访问 [go.dev](https://pkg.go.dev/github.com/xuri/xgen)。

`xgen` 命令可将 XML 模式定义文件编译为多语言类型或类声明的代码。

首先安装命令行工具:

```sh
go install github.com/xuri/xgen/cmd/xgen@latest
```

下面的命令将遍历 `xsd` 目录中的 XML 模式定义文件，并在 `output` 目录中生成 Go 语言结构体声明代码。

```text
$ xgen -i /path/to/your/xsd -o /path/to/your/output -l Go
```

Usage:

```text
$ xgen [<flag> ...] <XSD file or directory> ...
   -i <path> 指定存放 XML 模式代码文件的输入路径
   -o <path> 指定输出代码目录
   -p        指定生成代码所属包名称
   -l        指定生成类型或类声明代码语言类型 (Go/C/Java/Rust/TypeScript)
   -h        查看此帮助信息并退出
   -v        查看版本号并退出
```

## XSD (XML Schema Definition)

XSD 是万维网联盟 ([W3C](https://www.w3.org)) 推荐的标准，它指定了在可扩展标记语言 ([XML](https://www.w3.org/TR/xml/)) 文档中描述元素的规范。开发者可以使用它来验证文档中的每个项目内容，并可以检查它是否符合放置元素的说明。

XSD 是一种分离于 XML 本身的模式语言，可用于表示 XML 文档所必须遵循的一组规则，并可根据该规则进行模式有效性验证。

## 社区合作

欢迎您为此项目贡献代码，提出建议或问题、修复 Bug 以及参与讨论对新功能的想法。XML 符合标准：[XML Schema Part 1: Structures Second Edition](https://www.w3.org/TR/xmlschema-1/).

## 开源许可

本项目遵循 BSD 3-Clause 开源许可协议，访问 [https://opensource.org/licenses/BSD-3-Clause](https://opensource.org/licenses/BSD-3-Clause) 查看许可协议文件。

Logo 由 [xuri](https://xuri.me) 创作，遵循 [Creative Commons 3.0 Attributions license](http://creativecommons.org/licenses/by/3.0/) 创作共用授权条款。
