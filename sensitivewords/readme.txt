敏感词过滤模块
========

提供RESTful风格接口实现用户输入敏感词过滤

一、请求接口:
  1.接口: http://host:port/sensitive_filter/
  2.方法: POST
  3.请求body:
    (1)类型: JSON编码字符串
    (2)字段:
      struct {
        action 字符串: "query" | "add" | "remove" ---请求方法---
        text 字符串: ---用户输入字符串---
      }

二、响应:
  1.Content-Type: application/json
  2.响应body:
    (1)类型: JSON编码字符串
    (2)字段:
      struct {
        status 整数: ---详见状态码定义---
        errorDetail 字符串: ---错误详细描述, 如成功访问此字段为空---
        data 字符串数组: ---用户输入中所包含的敏感词如不包含敏感词则为空---
      }

三、算法及实现:
  从敏感词词库生成敏感词Trie树，收到用户输入后，将用户输入分片（例如：敏感词过滤, 将分片为[敏感词过滤， 感词过滤， 词过滤， 过滤, 滤]),
  针对每个分片生成一个goroutine进行匹配，然后汇总匹配结果返回

四、持久化:
  敏感词字典存储于mongodb中,服务器初始化时一次性加载全部敏感词在内存中生成Trie树，服务器收到add或remove请求时，先调用数据库接口修改数据库
  后修改Trie树，如修改数据库不成功则直接返回错误response，不对Trie树进行修改

五、返回状态码定义：
  200： 访问成功
  400： HTTP请求方法错误
  401: unmarshal json error
  402： 请求text为空
  500: 服务器端错误
