import matplotlib.pyplot as plt

# 数据
join_points = [1, 4, 8, 12, 16, 20, 24, 28, 32, 36]
wasm_times = [147, 540, 1078, 1687, 2238, 2943, 3503, 4112, 4724, 5209]
evm_times = [1702, 1717, 1791, 1892, 1983, 2144, 2260, 2254, 2306, 2352]

# 绘制曲线
plt.plot(join_points, wasm_times, label='wasm', marker='o')
plt.plot(join_points, evm_times, label='evm', marker='o')

# 添加标题和标签
plt.title('Execution Time vs. Join Point Count')
plt.xlabel('Join Point Count')
plt.ylabel('Execution Time (μs)')

# 添加图例
plt.legend()

plt.xticks(join_points)
# 显示图形
plt.grid(True)
plt.show()
