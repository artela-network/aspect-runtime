import matplotlib.pyplot as plt

# 数据
join_points = [1, 4, 8, 12, 16, 20, 24, 28, 32, 36]
wasm_times = [28, 111, 232, 361, 494, 640, 768, 926, 1112, 1276]
evm_times = [989, 1047, 1117, 1194, 1264, 1331, 1411, 1474, 1558, 1617]

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
