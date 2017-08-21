# harbor-tags-del
删除 harbor 里 registories 的 tags

```
projectID   = 3  // 项目ID
reserveDays = 16 // 分支保留天数
realDel     = false  // 仅打印全部分支，不会删除，设为 true 则执行删除；
```

主要是对 `docker registory` API 的操作。

注意：

通过 API 的方式删除分支，并不会实际删除存储在磁盘里的文件。要想真正达到清理磁盘的目的，请参考[Deleting repositories](https://github.com/vmware/harbor/blob/master/docs/user_guide.md#deleting-repositories)

