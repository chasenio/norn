1. 获取目标分支的最新 commit sha
2. 获取目标分支的 tree
3. 获取源分支的 commit
4. 将获取的 sha 和 tree 和 源分支的 commit 生成新的兄弟 commit， 和 new tree, 并强制更新临时分支
5. 将新的 commit merge 到临时分支
6. 使用 merge 后的 new Tree 创建新的 commit，
7. 使用 merge 后的 new commit 更新目标分支， 不强制， 保证是 fast-forward
