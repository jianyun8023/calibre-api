package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClearFieldsWithEmptyValues 测试清空字段的场景
// 规则：只有 Tags 可以清空，其他所有字段只允许补全（非空值）
func TestClearFieldsWithEmptyValues(t *testing.T) {
	// 模拟用户场景：书籍包含垃圾推广信息
	oldBook := &Book{
		ID:        274781,
		Title:     "多余人",
		Publisher: "(公众号：幸福的味道     不知道读什么书，就关注这个公众号)",
		Tags:      []string{"公众号：幸福的味道"},
		Authors:   []string{"屠格涅夫"},
		Comments:  "正常的书籍简介",
	}

	t.Run("清空标签 - 允许", func(t *testing.T) {
		// 只有 Tags 可以清空
		emptySlice := []string{}
		updates := &BookUpdate{
			Tags: emptySlice, // 清空标签
		}

		metadata := buildUpdateParams(updates, oldBook)

		// 验证标签被设置为空数组
		tags, tagsExists := metadata["tags"]
		assert.True(t, tagsExists, "tags 字段应该存在")
		assert.Equal(t, []string{}, tags, "tags 应该是空数组")
	})

	t.Run("清空出版社 - 不允许", func(t *testing.T) {
		// Publisher 不允许清空，只能补全
		emptyStr := ""
		updates := &BookUpdate{
			Publisher: &emptyStr, // 尝试清空出版社
		}

		metadata := buildUpdateParams(updates, oldBook)

		// 验证出版社不会被设置（空字符串被拒绝）
		_, pubExists := metadata["publisher"]
		assert.False(t, pubExists, "publisher 空字符串不应该被更新")
	})

	t.Run("清空评论 - 不允许", func(t *testing.T) {
		emptyStr := ""
		updates := &BookUpdate{
			Comments: &emptyStr, // 尝试清空评论
		}

		metadata := buildUpdateParams(updates, oldBook)

		_, exists := metadata["comments"]
		assert.False(t, exists, "comments 空字符串不应该被更新")
	})

	t.Run("清空作者 - 不允许", func(t *testing.T) {
		emptySlice := []string{}
		updates := &BookUpdate{
			Authors: emptySlice, // 尝试清空作者
		}

		metadata := buildUpdateParams(updates, oldBook)

		// 验证作者不会被设置（空数组被拒绝）
		_, exists := metadata["authors"]
		assert.False(t, exists, "authors 空数组不应该被更新")
	})

	t.Run("不更新字段 - nil 表示不更新", func(t *testing.T) {
		// 所有字段都是 nil
		updates := &BookUpdate{
			Tags:      nil, // nil 表示不更新
			Authors:   nil, // nil 表示不更新
			Publisher: nil, // nil 表示不更新
		}

		metadata := buildUpdateParams(updates, oldBook)

		// 验证 nil 字段不被设置
		_, tagsExists := metadata["tags"]
		assert.False(t, tagsExists, "tags 为 nil 时不应该更新")

		_, authorsExists := metadata["authors"]
		assert.False(t, authorsExists, "authors 为 nil 时不应该更新")

		_, pubExists := metadata["publisher"]
		assert.False(t, pubExists, "publisher 为 nil 时不应该更新")
	})

	t.Run("补全出版社信息 - 允许", func(t *testing.T) {
		// 更新出版社为非空值
		newPub := "人民文学出版社"
		updates := &BookUpdate{
			Publisher: &newPub, // 补全出版社
		}

		metadata := buildUpdateParams(updates, oldBook)

		// 验证出版社被更新
		publisher, pubExists := metadata["publisher"]
		assert.True(t, pubExists, "publisher 非空值应该被更新")
		assert.Equal(t, "人民文学出版社", publisher, "publisher 应该是新值")
	})
}

// TestUpdateWithValues 测试正常更新字段的场景
func TestUpdateWithValues(t *testing.T) {
	oldBook := &Book{
		ID:        1,
		Title:     "Old Title",
		Publisher: "Old Publisher",
		Tags:      []string{"old-tag"},
	}

	t.Run("更新多个字段", func(t *testing.T) {
		title := "New Title"
		publisher := "New Publisher"
		tags := []string{"new-tag"}
		authors := []string{"Author 1", "Author 2"}

		updates := &BookUpdate{
			Title:     &title,
			Publisher: &publisher,
			Tags:      tags,
			Authors:   authors,
		}

		metadata := buildUpdateParams(updates, oldBook)

		assert.Equal(t, "New Title", metadata["title"])
		assert.Equal(t, "New Publisher", metadata["publisher"])
		assert.Equal(t, []string{"new-tag"}, metadata["tags"])
		assert.Equal(t, []string{"Author 1", "Author 2"}, metadata["authors"])
	})
}
