scheduler

Quicker schedule generation for [classtime](https://github.com/rosshamish/classtime).

Feature roadmap:

- [x] represent Section as a concrete type (see rosshamish/classtime#103)
- [x] represent Schedule as a concrete type
- [ ] score schedules
	- [x] with len(Sections)
	- [ ] with rough ports of the Python scoring functions
	- [ ] with scoring functions whose scores are bell-curved (ish)
- [x] get section data from course ids
	- [x] directly from psql
	- [ ] from temporary classtime api endpoint, /api/vx/sections
	- [ ] ?
- [x] conflict detection
	- [x] same course, same component
	- [x] same time
	- [ ] course dependency not satisfied (ie LECA1->LABA2,3,4)
		- this is half-implemented, need to check the AutoEnroll component type, see the Python impl
- [x] gen with constraint solver (SAT solver)
- [ ] busy-times
- [ ] ~~more progressive conflict detection (see rosshamish/classtime#109)~~ SAT solver makes these unnecessary
	- [ ] ~~is this pair of sections a known conflict? -conflict~~
	- [ ] ~~are these sections on different days? -noconflict~~
	- [ ] ~~do their times overlap? -conflict~~
- [ ] conflict data collection
	- [ ] known conflicting section pairs per course pair
		-  "do these courses fit pretty well together? Or will this schedule be a lost cause with this pair?"
		- **this could lead to course suggestions based on what you've already got in your schedule**
	- [ ] return schedules with 1 or 2 conflicts but are great schedules
- [ ] caching
	- [ ] of known conflicting section pairs
	- [ ] of schedule requests
	- [ ] persistent across instances & restarts
- [ ] electives
- [ ] "more like this" condensing
- [ ] expose as an API with same interface as classtime/api/vN/generate-schedules
- [ ] redirect classtime/api/vN+1/generate-schedules requests to this API

Benchmark roadmap:

> comparing against ~30s for sequentially running all cases at [TestAPI.test_generate_schedules](https://github.com/rosshamish/classtime/blob/8236e91f001f4a5ba76bf1935055415784f2abfd/tests/classtime/test_api.py#L107)

- [x] running one test properly with bare minimum
- [ ] running properly with busy-times
- [ ] ~~running properly with electives too~~ it's broken in classtime right now anyway
- [ ] running all tests properly
- [ ] 90s
- [ ] 45s
- [ ] 30s
- [ ] 15s
- [ ] 10s
- [ ] 5s
- [ ] 1s
