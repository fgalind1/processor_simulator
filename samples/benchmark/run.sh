

rm -rf samples/benchmark/sample_loop
# bin/simulator.exe run samples/programs/sample_loop.txt -c samples/configs/nehalem_stall.config -o samples/benchmark/sample_loop/stall --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_loop.txt -c samples/configs/nehalem_always.config -o samples/benchmark/sample_loop/always --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_loop.txt -c samples/configs/nehalem_never.config -o samples/benchmark/sample_loop/never --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_loop.txt -c samples/configs/nehalem_backward.config -o samples/benchmark/sample_loop/backward --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_loop.txt -c samples/configs/nehalem_forward.config -o samples/benchmark/sample_loop/forward --max-cycles 2000
# rm -r samples/benchmark/*/*/debug.log
# rm -r samples/benchmark/*/*/assembly.hex



rm -rf samples/benchmark/sample_factorial
# bin/simulator.exe run samples/programs/sample_factorial.txt -c samples/configs/nehalem_stall.config -o samples/benchmark/sample_factorial/stall --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_factorial.txt -c samples/configs/nehalem_always.config -o samples/benchmark/sample_factorial/always --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_factorial.txt -c samples/configs/nehalem_never.config -o samples/benchmark/sample_factorial/never --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_factorial.txt -c samples/configs/nehalem_backward.config -o samples/benchmark/sample_factorial/backward --max-cycles 2000
# bin/simulator.exe run samples/programs/sample_factorial.txt -c samples/configs/nehalem_forward.config -o samples/benchmark/sample_factorial/forward --max-cycles 2000
# rm -r samples/benchmark/*/*/debug.log
# rm -r samples/benchmark/*/*/assembly.hex



rm -rf samples/benchmark/sample_bubble_sort
bin/simulator.exe run samples/programs/sample_bubble_sort.txt -c samples/configs/nehalem_stall.config -o samples/benchmark/sample_bubble_sort/stall --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort.txt -c samples/configs/nehalem_always.config -o samples/benchmark/sample_bubble_sort/always --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort.txt -c samples/configs/nehalem_never.config -o samples/benchmark/sample_bubble_sort/never --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort.txt -c samples/configs/nehalem_backward.config -o samples/benchmark/sample_bubble_sort/backward --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort.txt -c samples/configs/nehalem_forward.config -o samples/benchmark/sample_bubble_sort/forward --max-cycles 2000
rm -r samples/benchmark/*/*/debug.log
rm -r samples/benchmark/*/*/assembly.hex