package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// 配置常量
const (
	defaultBaseURL = "http://localhost:18060"
	httpTimeout    = 30 * time.Second
)

// 内部辅助函数：构建请求
func callAPI(ctx context.Context, method, path string, body map[string]interface{}) (map[string]interface{}, error) {
	client := http.DefaultClient

	var reqBody io.Reader
	if body != nil {
		jsonData, e := json.Marshal(body)
		if e != nil {
			return nil, fmt.Errorf("marshal body failed: %w", e)
}

		reqBody = jsonData
	}

	req, e := http.NewRequestWithContext(ctx, method, defaultBaseURL+path, reqBody)
	if e != nil {
		return nil, fmt.Errorf("new request failed: %w", e)
}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")

	resp, e := client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("http request failed: %w", e)
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return nil, fmt.Errorf("read response failed: %w", e)
}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error status %d: %s", resp.StatusCode, string(respBody))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", e)
}

	return result, nil
}

}

// 内部辅助函数：构建查询参数
func buildQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)

	return values.Encode()
}

}

// HandleLogin 登录和检查登录状态
func HandleLogin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// 检查登录状态
	status, apiErr := callAPI(ctx, "GET", "/api/check_login", nil)
	if apiErr != nil {
		// 如果检查失败，尝试执行登录流程
		_, loginErr := callAPI(ctx, "POST", "/api/login", nil)
		if loginErr != nil {
			return err(fmt.Sprintf("登录失败: %v", loginErr))
		}
		return ok("登录流程已启动，请在浏览器中完成扫码或验证")
	}

	if status["logged_in"] == true {
		return ok("已登录状态")
	}

	return ok("未登录，请调用登录功能")
}

// HandlePublishImageText 发布图文内容
func HandlePublishImageText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	imagesStr, _ :=getString(args, "images")
	tagsStr, _ :=getString(args, "tags")

	if title == "" {
		return err("标题不能为空")
	}
	if content == "" {
		return err("内容不能为空")
	}

	// 解析图片列表
	var images []string
	if imagesStr != "" {
		if e := json.Unmarshal([]byte(imagesStr), &images); e != nil {
			// 尝试按逗号分割
			images = strings.Split(imagesStr, ",")
			for i, img := range images {
				images[i] = strings.TrimSpace(img)

		}
	}

	// 解析标签列表
	var tags []string
	if tagsStr != "" {
		if e := json.Unmarshal([]byte(tagsStr), &tags); e != nil {
			tags = strings.Split(tagsStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)

		}
	}

	// 处理本地图片路径
	var processedImages []string
	for _, img := range images {
		if strings.HasPrefix(img, "/") || strings.HasPrefix(img, "C:\\") || strings.HasPrefix(img, "D:\\") {
			// 本地文件，需要上传或转换
			if _, e := os.Stat(img); os.IsNotExist(e) {
				return err(fmt.Sprintf("图片文件不存在: %s", img))
			}
			// 这里假设 API 支持直接处理本地路径，或者需要预先上传
			// 实际实现可能需要先上传文件获取 URL
			processedImages = append(processedImages, img)
		} else {
			processedImages = append(processedImages, img)

	}

	body := map[string]interface{}{
		"title":   title,
		"content": content,
		"images":  processedImages,
		"tags":    tags,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/publish/image_text", body)
	if apiErr != nil {
		return err(fmt.Sprintf("发布失败: %v", apiErr))
	}

	if result["success"] == true {
		return ok(fmt.Sprintf("发布成功，笔记ID: %v", result["note_id"]))
	}

	return err(fmt.Sprintf("发布失败: %v", result["message"]))
}

}
}
}

// HandlePublishVideo 发布视频内容
func HandlePublishVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	content, _ :=getString(args, "content")
	videoPath, _ :=getString(args, "video_path")
	tagsStr, _ :=getString(args, "tags")

	if title == "" {
		return err("标题不能为空")
	}
	if content == "" {
		return err("内容不能为空")
	}
	if videoPath == "" {
		return err("视频路径不能为空")
	}

	// 检查视频文件是否存在
	if _, e := os.Stat(videoPath); os.IsNotExist(e) {
		return err(fmt.Sprintf("视频文件不存在: %s", videoPath))
	}

	// 解析标签列表
	var tags []string
	if tagsStr != "" {
		if e := json.Unmarshal([]byte(tagsStr), &tags); e != nil {
			tags = strings.Split(tagsStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)

		}
	}

	body := map[string]interface{}{
		"title":   title,
		"content": content,
		"video":   videoPath,
		"tags":    tags,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/publish/video", body)
	if apiErr != nil {
		return err(fmt.Sprintf("视频发布失败: %v", apiErr))
	}

	if result["success"] == true {
		return ok(fmt.Sprintf("视频发布成功，笔记ID: %v (处理中)", result["note_id"]))
	}

	return err(fmt.Sprintf("视频发布失败: %v", result["message"]))
}

}

// HandleSearch 搜索内容
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	page, _ :=getInt(args, "page")

	if keyword == "" {
		return err("搜索关键词不能为空")
	}

	if page == 0 {
		page = 1
	}

	params := map[string]string{
		"keyword": keyword,
		"page":    strconv.Itoa(page),
	}

	query := buildQuery(params)
	result, apiErr := callAPI(ctx, "GET", "/api/search?"+query, nil)
	if apiErr != nil {
		return err(fmt.Sprintf("搜索失败: %v", apiErr))
	}

	// 格式化结果
	notes, _ := result["notes"].([]interface{})
	var noteList []string
	for _, note := range notes {
		if noteMap, found := note.(map[string]interface{}); found {
			title, _ := noteMap["title"].(string)
			noteID, _ := noteMap["note_id"].(string)
			noteList = append(noteList, fmt.Sprintf("ID: %s, 标题: %s", noteID, title))

	}

	return ok(fmt.Sprintf("搜索结果:\n%s", strings.Join(noteList, "\n")))
}

}

// HandleGetRecommendations 获取推荐列表
func HandleGetRecommendations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	page, _ :=getInt(args, "page")

	if page == 0 {
		page = 1
	}

	params := map[string]string{
		"page": strconv.Itoa(page),
	}

	query := buildQuery(params)
	result, apiErr := callAPI(ctx, "GET", "/api/recommendations?"+query, nil)
	if apiErr != nil {
		return err(fmt.Sprintf("获取推荐失败: %v", apiErr))
	}

	notes, _ := result["notes"].([]interface{})
	var noteList []string
	for _, note := range notes {
		if noteMap, found := note.(map[string]interface{}); found {
			title, _ := noteMap["title"].(string)
			noteID, _ := noteMap["note_id"].(string)
			noteList = append(noteList, fmt.Sprintf("ID: %s, 标题: %s", noteID, title))

	}

	return ok(fmt.Sprintf("推荐列表:\n%s", strings.Join(noteList, "\n")))
}

}

// HandleGetPostDetails 获取帖子详情
func HandleGetPostDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ :=getString(args, "feed_id")
	xsecToken, _ :=getString(args, "xsec_token")

	if feedID == "" {
		return err("feed_id 不能为空")
	}
	if xsecToken == "" {
		return err("xsec_token 不能为空")
	}

	body := map[string]interface{}{
		"feed_id":    feedID,
		"xsec_token": xsecToken,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/post/details", body)
	if apiErr != nil {
		return err(fmt.Sprintf("获取详情失败: %v", apiErr))
	}

	// 格式化返回信息
	title, _ := result["title"].(string)
	content, _ := result["content"].(string)
	likes, _ := result["likes"].(float64)
	favorites, _ := result["favorites"].(float64)
	comments, _ := result["comments"].(float64)

	return ok(fmt.Sprintf("帖子详情:\n标题: %s\n内容: %s\n点赞: %.0f\n收藏: %.0f\n评论: %.0f",
}
		title, content, likes, favorites, comments))
}

// HandlePostComment 发表评论到帖子
func HandlePostComment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ :=getString(args, "feed_id")
	xsecToken, _ :=getString(args, "xsec_token")
	comment, _ :=getString(args, "comment")

	if feedID == "" {
		return err("feed_id 不能为空")
	}
	if xsecToken == "" {
		return err("xsec_token 不能为空")
	}
	if comment == "" {
		return err("评论内容不能为空")
	}

	body := map[string]interface{}{
		"feed_id":    feedID,
		"xsec_token": xsecToken,
		"comment":    comment,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/comment/post", body)
	if apiErr != nil {
		return err(fmt.Sprintf("评论失败: %v", apiErr))
	}

	if result["success"] == true {
		return ok("评论发布成功")
	}

	return err(fmt.Sprintf("评论失败: %v", result["message"]))
}

// HandleGetUserProfile 获取用户个人主页
func HandleGetUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userID, _ :=getString(args, "user_id")
	xsecToken, _ :=getString(args, "xsec_token")

	if userID == "" {
		return err("user_id 不能为空")
	}
	if xsecToken == "" {
		return err("xsec_token 不能为空")
	}

	body := map[string]interface{}{
		"user_id":    userID,
		"xsec_token": xsecToken,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/user/profile", body)
	if apiErr != nil {
		return err(fmt.Sprintf("获取用户信息失败: %v", apiErr))
	}

	nickname, _ := result["nickname"].(string)
	bio, _ := result["bio"].(string)
	following, _ := result["following"].(float64)
	followers, _ := result["followers"].(float64)
	likes, _ := result["likes"].(float64)

	return ok(fmt.Sprintf("用户信息:\n昵称: %s\n简介: %s\n关注: %.0f\n粉丝: %.0f\n获赞: %.0f",
		nickname, bio, following, followers, likes))
}

// HandleReplyComment 回复评论
func HandleReplyComment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ :=getString(args, "feed_id")
	xsecToken, _ :=getString(args, "xsec_token")
	commentID, _ :=getString(args, "comment_id")
	userID, _ :=getString(args, "user_id")
	reply, _ :=getString(args, "reply")

	if feedID == "" {
		return err("feed_id 不能为空")
	}
	if xsecToken == "" {
		return err("xsec_token 不能为空")
	}
	if reply == "" {
		return err("回复内容不能为空")
	}
	if commentID == "" && userID == "" {
		return err("comment_id 或 user_id 至少提供一个")
	}

	body := map[string]interface{}{
		"feed_id":    feedID,
		"xsec_token": xsecToken,
		"reply":      reply,
	}
	if commentID != "" {
		body["comment_id"] = commentID
	}
	if userID != "" {
		body["user_id"] = userID
	}

	result, apiErr := callAPI(ctx, "POST", "/api/comment/reply", body)
	if apiErr != nil {
		return err(fmt.Sprintf("回复失败: %v", apiErr))
	}

	if result["success"] == true {
		return ok("回复成功")
	}

	return err(fmt.Sprintf("回复失败: %v", result["message"]))
}

// HandleLike 点赞/取消点赞
func HandleLike(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ :=getString(args, "feed_id")
	xsecToken, _ :=getString(args, "xsec_token")
	unlike, _ :=getBool(args, "unlike")

	if feedID == "" {
		return err("feed_id 不能为空")
	}
	if xsecToken == "" {
		return err("xsec_token 不能为空")
	}

	body := map[string]interface{}{
		"feed_id":    feedID,
		"xsec_token": xsecToken,
		"unlike":     unlike,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/like", body)
	if apiErr != nil {
		return err(fmt.Sprintf("点赞操作失败: %v", apiErr))
	}

	if result["success"] == true {
		action := "点赞"
		if unlike {
			action = "取消点赞"
		}
		return ok(fmt.Sprintf("%s成功", action))
	}

	return err(fmt.Sprintf("操作失败: %v", result["message"]))
}

// HandleFavorite 收藏/取消收藏
func HandleFavorite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ :=getString(args, "feed_id")
	xsecToken, _ :=getString(args, "xsec_token")
	unfavorite, _ :=getBool(args, "unfavorite")

	if feedID == "" {
		return err("feed_id 不能为空")
	}
	if xsecToken == "" {
		return err("xsec_token 不能为空")
	}

	body := map[string]interface{}{
		"feed_id":    feedID,
		"xsec_token": xsecToken,
		"unfavorite": unfavorite,
	}

	result, apiErr := callAPI(ctx, "POST", "/api/favorite", body)
	if apiErr != nil {
		return err(fmt.Sprintf("收藏操作失败: %v", apiErr))
	}

	if result["success"] == true {
		action := "收藏"
		if unfavorite {
			action = "取消收藏"
		}
		return ok(fmt.Sprintf("%s成功", action))
	}

	return err(fmt.Sprintf("操作失败: %v", result["message"]))
}