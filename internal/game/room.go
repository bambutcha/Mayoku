package game

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Chelaran/mayoku/internal/models"
	logger "github.com/Chelaran/yagalog"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Room представляет игровую комнату
type Room struct {
	mu      sync.RWMutex
	state   *RoomState
	clients map[uint]*Client // user_id -> Client
	db      *gorm.DB
	redis   *redis.Client
	log     *logger.Logger
	timer   *time.Timer
}

// NewRoom создает новую комнату
func NewRoom(roomID string, createdBy uint, deckID uint, deckName string, maxPlayers, spyCount, duration int, db *gorm.DB, redis *redis.Client) *Room {
	log, _ := logger.NewLogger()
	room := &Room{
		state: &RoomState{
			RoomID:     roomID,
			Status:     StatusWaiting,
			Players:    make(map[uint]*Player),
			DeckID:     deckID,
			DeckName:   deckName,
			MaxPlayers: maxPlayers,
			SpyCount:   spyCount,
			Duration:   duration,
			CreatedBy:  createdBy,
			CreatedAt:  time.Now(),
		},
		clients: make(map[uint]*Client),
		db:      db,
		redis:   redis,
		log:     log,
	}

	// Сохраняем в Redis
	room.saveToRedis()

	return room
}

// IsRoomAdmin проверяет, является ли пользователь админом комнаты (создателем)
func (r *Room) IsRoomAdmin(userID uint) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.CreatedBy == userID
}

// AddPlayer добавляет игрока в комнату
func (r *Room) AddPlayer(userID uint, tgID int64, username, avatarURL string, client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем лимит игроков
	if len(r.state.Players) >= r.state.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	// Проверяем, не добавлен ли уже
	if _, exists := r.state.Players[userID]; exists {
		return fmt.Errorf("player already in room")
	}

	// Добавляем игрока
	r.state.Players[userID] = &Player{
		UserID:    userID,
		TgID:      tgID,
		Username:  username,
		AvatarURL: avatarURL,
		IsReady:   false,
	}

	r.clients[userID] = client

	// Сохраняем в Redis
	r.saveToRedis()

	// Отправляем обновление всем
	r.broadcastState()

	return nil
}

// RemovePlayer удаляет игрока из комнаты
func (r *Room) RemovePlayer(userID uint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.state.Players, userID)
	delete(r.clients, userID)

	// Если комната пуста, можно удалить
	if len(r.state.Players) == 0 {
		r.state.Status = StatusFinished
	}

	r.saveToRedis()
	r.broadcastState()
}

// KickPlayer удаляет игрока из комнаты (только админ комнаты)
func (r *Room) KickPlayer(adminUserID, targetUserID uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем права админа комнаты
	if r.state.CreatedBy != adminUserID {
		return fmt.Errorf("only room admin can kick players")
	}

	// Нельзя выгнать самого себя
	if adminUserID == targetUserID {
		return fmt.Errorf("cannot kick yourself")
	}

	// Проверяем, что игрок существует
	if _, exists := r.state.Players[targetUserID]; !exists {
		return fmt.Errorf("player not found in room")
	}

	// Удаляем игрока
	delete(r.state.Players, targetUserID)
	if client, exists := r.clients[targetUserID]; exists {
		delete(r.clients, targetUserID)
		// Отправляем уведомление выгнанному игроку
		msg := WSMessage{
			Type: "kicked_from_room",
			Payload: map[string]interface{}{
				"room_id": r.state.RoomID,
				"reason":  "kicked by room admin",
			},
		}
		client.SendMessage(msg)
	}

	r.saveToRedis()
	r.broadcastState()

	return nil
}

// UpdateSettings обновляет настройки комнаты (только админ комнаты)
func (r *Room) UpdateSettings(adminUserID uint, maxPlayers, spyCount, duration *int, deckID *uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем права админа комнаты
	if r.state.CreatedBy != adminUserID {
		return fmt.Errorf("only room admin can update settings")
	}

	// Проверяем, что игра еще не началась
	if r.state.Status != StatusWaiting {
		return fmt.Errorf("cannot update settings: game already started")
	}

	// Обновляем настройки
	if maxPlayers != nil {
		if *maxPlayers < 3 || *maxPlayers > 10 {
			return fmt.Errorf("max_players must be between 3 and 10")
		}
		r.state.MaxPlayers = *maxPlayers
	}

	if spyCount != nil {
		if *spyCount < 1 || *spyCount > 2 {
			return fmt.Errorf("spy_count must be 1 or 2")
		}
		r.state.SpyCount = *spyCount
	}

	if duration != nil {
		if *duration < 3 || *duration > 15 {
			return fmt.Errorf("duration must be between 3 and 15 minutes")
		}
		r.state.Duration = *duration
	}

	if deckID != nil {
		// Проверяем существование колоды
		var deck models.Deck
		if err := r.db.First(&deck, *deckID).Error; err != nil {
			return fmt.Errorf("deck not found")
		}
		r.state.DeckID = *deckID
		r.state.DeckName = deck.Name
	}

	r.saveToRedis()
	r.broadcastState()

	return nil
}

// SetPlayerReady устанавливает готовность игрока
func (r *Room) SetPlayerReady(userID uint, ready bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player, exists := r.state.Players[userID]
	if !exists {
		return fmt.Errorf("player not found")
	}

	player.IsReady = ready
	r.saveToRedis()
	r.broadcastState()

	// Проверяем, все ли готовы (минимум 3 игрока)
	if ready && len(r.state.Players) >= 3 {
		allReady := true
		for _, p := range r.state.Players {
			if !p.IsReady {
				allReady = false
				break
			}
		}
		if allReady {
			go r.startGame()
		}
	}

	return nil
}

// startGame начинает игру
func (r *Room) startGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state.Status != StatusWaiting {
		return
	}

	// Загружаем локации из колоды
	var locations []models.Location
	if err := r.db.Where("deck_id = ?", r.state.DeckID).Find(&locations).Error; err != nil {
		r.log.Error("Failed to load locations: %v", err)
		return
	}

	if len(locations) == 0 {
		r.log.Error("No locations in deck")
		return
	}

	// Выбираем случайную локацию
	rand.Seed(time.Now().UnixNano())
	location := locations[rand.Intn(len(locations))]

	// Сохраняем информацию о локации
	r.state.Location = &LocationInfo{
		Name:     location.Name,
		ImageURL: location.ImageURL,
		Roles:    []string(location.Roles),
	}

	// Раздаем роли
	playerIDs := make([]uint, 0, len(r.state.Players))
	for id := range r.state.Players {
		playerIDs = append(playerIDs, id)
	}

	// Перемешиваем
	rand.Shuffle(len(playerIDs), func(i, j int) {
		playerIDs[i], playerIDs[j] = playerIDs[j], playerIDs[i]
	})

	// Назначаем шпионов
	spyCount := r.state.SpyCount
	if spyCount > len(playerIDs) {
		spyCount = len(playerIDs) / 2 // Максимум половина
	}
	if spyCount == 0 {
		spyCount = 1
	}

	r.state.SpyIDs = playerIDs[:spyCount]
	spyMap := make(map[uint]bool)
	for _, id := range r.state.SpyIDs {
		spyMap[id] = true
	}

	// Назначаем роли и локации
	roleIndex := 0
	for _, playerID := range playerIDs {
		player := r.state.Players[playerID]
		if spyMap[playerID] {
			player.Role = RoleSpy
		} else {
			player.Role = RoleLocal
			player.Location = location.Name
			if roleIndex < len(location.Roles) {
				player.LocationRole = location.Roles[roleIndex]
				roleIndex++
			}
		}
	}

	// Устанавливаем таймер
	duration := time.Duration(r.state.Duration) * time.Minute
	timerEnd := time.Now().Add(duration)
	r.state.TimerEnd = &timerEnd
	r.state.Status = StatusPlaying

	// Запускаем таймер
	r.timer = time.AfterFunc(duration, func() {
		r.handleTimerExpired()
	})

	r.saveToRedis()

	// Отправляем каждому игроку его роль
	r.sendRolesToPlayers()
}

// sendRolesToPlayers отправляет каждому игроку его роль
func (r *Room) sendRolesToPlayers() {
	for userID, player := range r.state.Players {
		client, exists := r.clients[userID]
		if !exists {
			continue
		}

		// Создаем персональное сообщение
		personalState := map[string]interface{}{
			"role": player.Role,
		}

		if player.Role == RoleLocal {
			personalState["location"] = player.Location
			personalState["location_role"] = player.LocationRole
		}

		msg := WSMessage{
			Type: "game_started",
			Payload: map[string]interface{}{
				"room_id":   r.state.RoomID,
				"my_role":   personalState,
				"timer_end": r.state.TimerEnd.Unix(),
				"spy_count": len(r.state.SpyIDs),
			},
		}

		client.SendMessage(msg)
	}

	// Отправляем общее обновление состояния
	r.broadcastState()
}

// handleTimerExpired обрабатывает истечение таймера
func (r *Room) handleTimerExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state.Status != StatusPlaying {
		return
	}

	// Победа шпиона (таймер истек)
	r.state.Winner = "spy"
	r.state.Status = StatusFinished
	r.finishGame()
}

// StartVoting начинает голосование
func (r *Room) StartVoting(initiatorID, targetUserID uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state.Status != StatusPlaying {
		return fmt.Errorf("game is not in playing status")
	}

	// Проверяем, что цель существует
	if _, exists := r.state.Players[targetUserID]; !exists {
		return fmt.Errorf("target player not found")
	}

	// Сбрасываем предыдущее голосование
	r.state.Voting = &VotingState{
		TargetUserID: targetUserID,
		Votes:        make(map[uint]bool),
		StartedAt:    time.Now(),
	}

	// Сбрасываем флаги голосования
	for _, player := range r.state.Players {
		player.IsVoted = false
	}

	r.state.Status = StatusVoting
	r.saveToRedis()

	// Отправляем уведомление о голосовании
	msg := WSMessage{
		Type: "vote_initiated",
		Payload: map[string]interface{}{
			"target_user_id": targetUserID,
			"initiator_id":   initiatorID,
		},
	}
	r.broadcastMessage(msg)

	return nil
}

// Vote обрабатывает голос игрока
func (r *Room) Vote(userID uint, vote bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state.Status != StatusVoting {
		return fmt.Errorf("no active voting")
	}

	if r.state.Voting == nil {
		return fmt.Errorf("voting not initialized")
	}

	// Нельзя голосовать за себя
	if userID == r.state.Voting.TargetUserID {
		return fmt.Errorf("cannot vote for yourself")
	}

	player, exists := r.state.Players[userID]
	if !exists {
		return fmt.Errorf("player not found")
	}

	if player.IsVoted {
		return fmt.Errorf("already voted")
	}

	// Записываем голос
	player.IsVoted = true
	player.Vote = vote
	r.state.Voting.Votes[userID] = vote

	r.saveToRedis()
	r.broadcastState()

	// Проверяем, все ли проголосовали
	allVoted := true
	for id, p := range r.state.Players {
		if id != r.state.Voting.TargetUserID && !p.IsVoted {
			allVoted = false
			break
		}
	}

	if allVoted {
		r.processVotingResult()
	}

	return nil
}

// processVotingResult обрабатывает результат голосования
func (r *Room) processVotingResult() {
	// Подсчитываем голоса "за"
	votesFor := 0
	for _, vote := range r.state.Voting.Votes {
		if vote {
			votesFor++
		}
	}

	// Нужно единогласие (все кроме обвиняемого)
	requiredVotes := len(r.state.Players) - 1

	if votesFor == requiredVotes {
		// Единогласное голосование - проверяем роль
		targetPlayer := r.state.Players[r.state.Voting.TargetUserID]
		if targetPlayer.Role == RoleSpy {
			// Победа местных
			r.state.Winner = "locals"
			r.state.Status = StatusFinished
			r.finishGame()
		} else {
			// Ошиблись - победа шпиона
			r.state.Winner = "spy"
			r.state.Status = StatusFinished
			r.finishGame()
		}
	} else {
		// Не единогласие - продолжаем игру
		r.state.Voting = nil
		r.state.Status = StatusPlaying
		r.broadcastState()
	}
}

// SpyGuess обрабатывает попытку шпиона угадать локацию
func (r *Room) SpyGuess(userID uint, locationName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state.Status != StatusPlaying {
		return fmt.Errorf("game is not in playing status")
	}

	player, exists := r.state.Players[userID]
	if !exists {
		return fmt.Errorf("player not found")
	}

	if player.Role != RoleSpy {
		return fmt.Errorf("only spy can guess location")
	}

	// Проверяем угадал ли
	guessed := locationName == r.state.Location.Name

	if guessed {
		r.state.Winner = "spy"
	} else {
		r.state.Winner = "locals"
	}

	r.state.Status = StatusFinished
	r.finishGame()

	return nil
}

// finishGame завершает игру
func (r *Room) finishGame() {
	if r.timer != nil {
		r.timer.Stop()
	}

	// Сохраняем в GameHistory
	go r.saveGameHistory()

	// Отправляем результаты
	msg := WSMessage{
		Type: "game_over",
		Payload: map[string]interface{}{
			"winner":   r.state.Winner,
			"spy_ids":  r.state.SpyIDs,
			"location": r.state.Location,
		},
	}
	r.broadcastMessage(msg)

	r.saveToRedis()
}

// saveGameHistory сохраняет историю игры в БД
func (r *Room) saveGameHistory() {
	duration := int(time.Since(r.state.CreatedAt).Seconds())

	history := models.GameHistory{
		RoomUUID: r.state.RoomID,
		DeckName: r.state.DeckName,
		Winner:   r.state.Winner,
		Duration: duration,
	}

	// Получаем пользователей
	var userIDs []uint
	for id := range r.state.Players {
		userIDs = append(userIDs, id)
	}

	var users []models.User
	r.db.Where("id IN ?", userIDs).Find(&users)
	history.Players = users

	r.db.Create(&history)

	// Обновляем статистику пользователей
	for _, player := range r.state.Players {
		var user models.User
		if err := r.db.First(&user, player.UserID).Error; err != nil {
			continue
		}

		user.GamesPlayed++

		isWinner := (r.state.Winner == "spy" && player.Role == RoleSpy) ||
			(r.state.Winner == "locals" && player.Role == RoleLocal)

		if isWinner {
			if player.Role == RoleSpy {
				user.WinsSpy++
			} else {
				user.WinsLocal++
			}
		} else {
			if player.Role == RoleSpy {
				user.LossesSpy++
			} else {
				user.LossesLocal++
			}
		}

		r.db.Save(&user)
	}
}

// broadcastState отправляет текущее состояние всем клиентам
func (r *Room) broadcastState() {
	msg := WSMessage{
		Type:    "room_update",
		Payload: r.getPublicState(),
	}
	r.broadcastMessage(msg)
}

// broadcastMessage отправляет сообщение всем клиентам
func (r *Room) broadcastMessage(msg WSMessage) {
	data, _ := json.Marshal(msg)
	for _, client := range r.clients {
		client.SendRaw(data)
	}
}

// getPublicState возвращает публичное состояние (без скрытых ролей)
func (r *Room) getPublicState() map[string]interface{} {
	players := make([]map[string]interface{}, 0, len(r.state.Players))
	for _, player := range r.state.Players {
		p := map[string]interface{}{
			"user_id":    player.UserID,
			"tg_id":      player.TgID,
			"username":   player.Username,
			"avatar_url": player.AvatarURL,
			"is_ready":   player.IsReady,
		}

		// Роль показываем только после окончания игры
		if r.state.Status == StatusFinished {
			p["role"] = player.Role
		}

		players = append(players, p)
	}

	state := map[string]interface{}{
		"room_id":     r.state.RoomID,
		"status":      r.state.Status,
		"players":     players,
		"max_players": r.state.MaxPlayers,
		"deck_name":   r.state.DeckName,
		"created_by":  r.state.CreatedBy, // ID создателя комнаты
	}

	if r.state.TimerEnd != nil {
		state["timer_end"] = r.state.TimerEnd.Unix()
	}

	if r.state.Voting != nil {
		state["voting"] = map[string]interface{}{
			"target_user_id": r.state.Voting.TargetUserID,
			"votes":          r.state.Voting.Votes,
		}
	}

	if r.state.Status == StatusFinished {
		state["winner"] = r.state.Winner
		state["spy_ids"] = r.state.SpyIDs
		state["location"] = r.state.Location
	}

	return state
}

// saveToRedis сохраняет состояние в Redis
func (r *Room) saveToRedis() {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s", r.state.RoomID)

	data, err := json.Marshal(r.state)
	if err != nil {
		r.log.Error("Failed to marshal room state: %v", err)
		return
	}

	r.redis.Set(ctx, key, data, time.Hour)
}

// LoadFromRedis загружает состояние из Redis
func (r *Room) LoadFromRedis() error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s", r.state.RoomID)

	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, r.state)
}
