import matplotlib.pyplot as plt

# 数据
add_counts = [0, 400, 800, 1200, 1600, 2000, 2400, 2800, 3200, 3600, 4000]
wasm_times = [1318, 1426, 1389, 1421, 1411, 1438, 1448, 1470, 1490, 1521, 1509]
evm_times = [259, 501, 775, 1031, 1303, 1587, 1913, 2159, 2513, 2760, 2904]

# 绘制对比图
plt.plot(add_counts, wasm_times, label='wasm', marker='o')
plt.plot(add_counts, evm_times, label='evm', marker='o')

# 添加标题和标签
plt.title('Execution Time vs. Computational Capacity')
plt.xlabel('Total Count of Adds (Addition Operations)')
plt.ylabel('Execution Time (μs)')

# 添加图例
plt.legend()

plt.xticks(add_counts)

# 显示图形
plt.grid(True)
plt.show()


