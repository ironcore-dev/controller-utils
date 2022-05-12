module example.org/testmod1

go 1.17

replace example.org/testmod2 => ./../testmod2

require example.org/testmod2 v0.0.0-00010101000000-000000000000
