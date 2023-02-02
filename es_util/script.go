package g_es

const EsIncrease = "ctx._source['%s'] += '%d'"

const EsArrAdd = "ctx._source['%s'].add('%s')"

const EsArrRemove = "ctx._source['%s'].remove(ctx._source['%s'].indexOf('%s'))"
