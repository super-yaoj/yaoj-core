/*
Package data provides universal control of data.

初代的 workflow 设计是将所有 processor 的 IO 都限定为文件，这样的好处是统一了交互的标准，但缺点
同样明显，不易扩展。而这也间接导致题目的底层数据存储混乱，缺失了静态类型的保护，有很大潜在问题。

Package data 目标是使 processor 不再仅限于文件读入，以更灵活的方式读入信息不仅可以提高效率，同时
可以极大地简化不必要的内容转换，也为缓存等提供更好的设施。
*/
package data
