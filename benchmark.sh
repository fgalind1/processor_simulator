#! /bin/bash

# Run benchmarks
bin/simulator.exe run samples/programs/sample_bubble_sort1.txt -c samples/configs/nehalem_stall.config -o benchmark/bubble_sort1/stall --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort1.txt -c samples/configs/nehalem_always.config -o benchmark/bubble_sort1/always --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort1.txt -c samples/configs/nehalem_never.config -o benchmark/bubble_sort1/never --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort1.txt -c samples/configs/nehalem_backward.config -o benchmark/bubble_sort1/backward --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort1.txt -c samples/configs/nehalem_forward.config -o benchmark/bubble_sort1/forward --max-cycles 2000

bin/simulator.exe run samples/programs/sample_bubble_sort2.txt -c samples/configs/nehalem_stall.config -o benchmark/bubble_sort2/stall --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort2.txt -c samples/configs/nehalem_always.config -o benchmark/bubble_sort2/always --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort2.txt -c samples/configs/nehalem_never.config -o benchmark/bubble_sort2/never --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort2.txt -c samples/configs/nehalem_backward.config -o benchmark/bubble_sort2/backward --max-cycles 2000
bin/simulator.exe run samples/programs/sample_bubble_sort2.txt -c samples/configs/nehalem_forward.config -o benchmark/bubble_sort2/forward --max-cycles 2000


# Merge output files
target_dir_path="benchmark/stats"
mkdir "${target_dir_path}"

for file in benchmark/*/*/output.log; do
        program="${file#*/}"
        program="${program%%/*}"

        config="${file#*/}"
        config="${config#*/}"
        config="${config%%/*}"

        target_file_name="${program}_${config}.log"
        cp "$file" "${target_dir_path}/${target_file_name}"
done