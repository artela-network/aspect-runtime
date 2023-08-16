import matplotlib.pyplot as plt

# 数据
add_counts = [0, 400, 800, 1200, 1600, 2000, 2400, 2800, 3200, 3600]
wasm_times = [309, 312, 313, 324, 323, 333, 339, 340, 347, 349]
evm_times = [192, 427, 669, 910, 1153, 1395, 1639, 1888, 2142, 2383]

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


