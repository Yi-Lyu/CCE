#!/bin/bash
# 快速部署脚本 - 推送干净的代码到GitHub

echo "🚀 Claude Code Exchange - 安全部署脚本"
echo "========================================"
echo ""

# 检查是否在正确的目录
if [ ! -f "proxy/go.mod" ]; then
    echo "❌ 错误：请在项目根目录运行此脚本"
    exit 1
fi

# 显示当前状态
echo "📊 当前Git状态："
git status --short
echo ""

# 确认没有敏感信息
echo "🔍 检查敏感信息..."
if grep -r "76ff9d11\|bccd0a1879\|ilywkmL5tw7npNiG\|asia.qcode.cc\|103.218.243" . 2>/dev/null | grep -v deploy.sh; then
    echo "❌ 发现敏感信息！请先清理后再部署"
    exit 1
fi
echo "✅ 没有发现敏感信息"
echo ""

# 获取用户输入
echo "请选择部署方式："
echo "1) 创建新的GitHub仓库"
echo "2) 强制替换现有仓库（危险！）"
echo "3) 查看当前远程仓库"
echo "4) 退出"
echo ""
read -p "请输入选项 (1-4): " choice

case $choice in
    1)
        echo ""
        read -p "请输入GitHub用户名: " username
        read -p "请输入仓库名称 (默认: Claude-Code-Exchange): " repo
        repo=${repo:-Claude-Code-Exchange}

        echo ""
        echo "将要执行的命令："
        echo "  git remote add origin https://github.com/$username/$repo.git"
        echo "  git branch -M main"
        echo "  git push -u origin main"
        echo ""
        read -p "确认执行？(y/n): " confirm

        if [ "$confirm" = "y" ]; then
            git remote add origin "https://github.com/$username/$repo.git" 2>/dev/null || {
                echo "远程仓库已存在，更新URL..."
                git remote set-url origin "https://github.com/$username/$repo.git"
            }
            git branch -M main
            git push -u origin main
            echo "✅ 推送完成！"
            echo "🔗 访问: https://github.com/$username/$repo"
        fi
        ;;

    2)
        echo ""
        echo "⚠️  警告：这将删除远程仓库的所有历史记录！"
        echo ""
        read -p "请输入GitHub用户名: " username
        read -p "请输入仓库名称: " repo

        echo ""
        echo "将要执行的命令："
        echo "  git remote set-url origin https://github.com/$username/$repo.git"
        echo "  git push --force origin main"
        echo ""
        read -p "确认强制推送？输入 'DELETE HISTORY' 确认: " confirm

        if [ "$confirm" = "DELETE HISTORY" ]; then
            git remote add origin "https://github.com/$username/$repo.git" 2>/dev/null || {
                git remote set-url origin "https://github.com/$username/$repo.git"
            }
            git push --force origin main
            echo "✅ 强制推送完成！"

            # 询问是否删除其他分支
            read -p "是否删除远程的其他分支？(y/n): " delbranch
            if [ "$delbranch" = "y" ]; then
                git push origin --delete feature/macos-client 2>/dev/null || echo "分支 feature/macos-client 不存在或已删除"
                git push origin --delete feature/intelligent-proxy 2>/dev/null || echo "分支 feature/intelligent-proxy 不存在或已删除"
                git push origin --delete feature/clean-config 2>/dev/null || echo "分支 feature/clean-config 不存在或已删除"
            fi
        else
            echo "❌ 操作已取消"
        fi
        ;;

    3)
        echo ""
        echo "当前远程仓库："
        git remote -v
        ;;

    4)
        echo "退出"
        exit 0
        ;;

    *)
        echo "❌ 无效选项"
        exit 1
        ;;
esac

echo ""
echo "📝 后续步骤："
echo "1. 复制您的本地配置: cp ~/code/Claude-Code-Exchange/proxy/configs/config.local.yaml ./proxy/configs/"
echo "2. 访问GitHub确认代码已推送"
echo "3. 考虑添加LICENSE文件"
echo ""
echo "✅ 完成！"