scheduler

Quicker schedule generation for [classtime](https://github.com/rosshamish/classtime).

Feature roadmap:

- [x] represent Section as a concrete type (see rosshamish/classtime#103)
- [x] represent Schedule as a concrete type
- [ ] score schedules
	- [ ] with rough ports of the Python scoring functions
	- [ ] with scoring functions whose scores are bell-curved (ish)
- [ ] conflict detection (see rosshamish/classtime#109)
- [ ] caching
	- [ ] of conflicts
	- [ ] of requests
	- [ ] persistent across instances & restarts
- [ ] electives
- [ ] "more like this" condensing
- [ ] conflict "error reporting" (see rosshamish/classtime#109)
- [ ] expose as an API with same interface as classtime/api/vN/generate-schedules
- [ ] redirect classtime/api/vN+1/generate-schedules requests to this API

Benchmark roadmap:

> comparing against ~30s for sequentially running all cases at [TestAPI.test_generate_schedules](https://github.com/rosshamish/classtime/blob/8236e91f001f4a5ba76bf1935055415784f2abfd/tests/classtime/test_api.py#L107)

- [ ] running properly with bare minimum
- [ ] running properly with busy-times
- [ ] ~running properly with electives too~ it's broken in classtime right now anyway
- [ ] 90s
- [ ] 45s
- [ ] 30s
- [ ] 15s
- [ ] 10s
- [ ] 5s
