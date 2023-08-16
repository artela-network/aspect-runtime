import matplotlib.pyplot as plt

# 数据
join_points = [1, 4, 8, 12, 16, 20, 24, 28, 32, 36, 37]
wasm_add_counts = [4045000, 3565000, 3075000, 2675000, 2275000, 1890000, 1500000, 1050000, 670000, 106000, 0]
evm_add_counts = [6020, 5830, 5780, 5660, 5540, 5480, 5430, 5320, 5180, 5030, 5020]

# 绘制对比图
plt.plot(join_points, wasm_add_counts, label='wasm', marker='o')
plt.plot(join_points, evm_add_counts, label='evm', marker='o')

# 添加标题和标签
plt.title('Computational Capacity vs. Join Point Count')
plt.xlabel('Join Point Count')
plt.ylabel('Total Count of Adds (Addition Operations)')

# 添加图例
plt.legend()

join_points1 = [1, 4, 8, 12, 16, 20, 24, 28, 32, 36]
plt.xticks(join_points1)

# 显示图形
plt.grid(True)
plt.show()


