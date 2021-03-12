package config

import scaffold "github.com/moetang/webapp-scaffold"

var webscaffold *scaffold.WebappScaffold

func InitWebScaffold(scaffold *scaffold.WebappScaffold) {
	webscaffold = scaffold
}
