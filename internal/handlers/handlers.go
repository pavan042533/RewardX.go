package handlers

import (
	"authapi/internal/db"
	"authapi/internal/models"
	"authapi/internal/utils"
	"strconv"
	"sync"
	"time"
	"github.com/gofiber/fiber/v2"
)

// RegisterUser handles user registration
func RegisterUser(c *fiber.Ctx) error {
	var u models.User
	if err := c.BodyParser(&u); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	if u.Role != "partner" && u.Role != "user" {
		u.Role = "user"
	}
	// checking if the user exists
    var existing models.User
	result := db.DB.Where("email = ?", u.Email).First(&existing)
	if result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already Registered",
		})
	}
	u.Points = 400
	// password hasshing
	hashedPassword, err := utils.HashingPassword(u.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}
	u.Password = hashedPassword
    // generating the otp
	u.OTP = utils.GenerateOTP()
	u.OTPExpiresAt = time.Now().Add(5 * time.Minute)
    // saves New user 
	if err := db.DB.Create(&u).Error; err!=nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not save user"})
	}
	//  Send OTP to email (mocked)
	if errmail := utils.SendOTPEmail(u.Email, u.OTP); errmail!=nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not send otp to user"})
	}

	//  Return success message
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered. OTP sent to email",
	})
}

// VerifyOTP handles user OTP verification
func VerifyOTP(c *fiber.Ctx) error {
    var input struct {
        Email    string `json:"email"`
        OTP      string `json:"otp"`
    }
    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }
    if input.Email == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Email required"})
    }
    var user models.User
    result := db.DB.Where("email = ?",  input.Email).First(&user)
    if result.Error != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }
    if !user.IsVerified && time.Now().After(user.OTPExpiresAt) {
        return c.Status(400).JSON(fiber.Map{"error": "OTP expired"})
    }
    if user.OTP != input.OTP {
        return c.Status(400).JSON(fiber.Map{"error": "Incorrect OTP"})
    }
    user.IsVerified = true
    db.DB.Save(&user)
    return c.JSON(fiber.Map{"message": "Verification successful"})
}

// LoginHandler handles user login and token generation
func LoginHandler(c *fiber.Ctx) error {
    var credentials struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.BodyParser(&credentials); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }
    if credentials.Email == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username or Email required"})
    }
    var user models.User
    result := db.DB.Where("email = ?", credentials.Email).First(&user)
    if result.Error != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Account Not found"})
    }
    if  !utils.CheckPasswordHashing(credentials.Password, user.Password) { 
        return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
    }
    if !user.IsVerified {
        return c.Status(403).JSON(fiber.Map{"error": "Account not verified"})
    }
	
    token, err := utils.GenerateToken(user.ID, user.Role)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
    }
    return c.JSON(fiber.Map{"token": token, "role": user.Role})
}

// ListRewards retrieves all available rewards
func ListRewards(c *fiber.Ctx) error {
	rewards:= []models.Reward{}
	if result:= db.DB.Find(&rewards); result.Error!=nil{
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Could not find Rewards"})
	}
	return c.JSON(rewards)
}

// GetUserWallet retrieves the points of the logged-in user
func GetUserWallet(c*fiber.Ctx) error {
	userID:= uint(c.Locals("user_id").(float64))
	var user models.User
	if result:= db.DB.Where("id = ?", userID).First(&user); result.Error!=nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(fiber.Map{"points": user.Points})
}

// RedeemReward allows a user to redeem a reward
func RedeemReward(c *fiber.Ctx) error{
	userID:= uint(c.Locals("user_id").(float64))
	var t models.Transaction
	if err:= c.BodyParser(&t); err!=nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user   models.User
	var reward models.Reward
	var wg     sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		db.DB.First(&user, userID)
	}()
	go func() {
		defer wg.Done()
		db.DB.First(&reward, t.RewardID)
	}()
	wg.Wait()
	if user.Points<reward.Cost{
		return c.Status(400).JSON(fiber.Map{"error": "Insuficient Points"})
	}else if reward.Stock<=0 {
		return  c.Status(400).JSON(fiber.Map{"error": "Reward out of stock"})
	}
	t.UserID = userID
	user.Points-=reward.Cost
	reward.Stock-=1
	t.Status = "Completed"
	t.PointsUsed = reward.Cost
	t.CouponCode = utils.GenerateCouponCode(reward.Name[:3])
	t.CreatedAt = time.Now()
	db.DB.Save(&user)
	db.DB.Save(&reward)
	db.DB.Create(&t)
	return c.JSON(fiber.Map{"message": "Reward redeemed", "transaction": t})
}

// GetUserTransactions retrieves all transactions for the logged-in user
func GetUserTransactions(c *fiber.Ctx) error {
	userID:=uint(c.Locals("user_id").(float64))
	var transactions []models.Transaction
	db.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&transactions)
	return c.JSON(transactions)
}

// AdminAddReward allows the admin to add a new reward
func AdminAddReward(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	userID:= uint(c.Locals("user_id").(float64))
	if role!="admin"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	var reward models.Reward
	if err:=c.BodyParser(&reward); err!=nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	reward.CreatedByID = userID
	db.DB.Create(&reward)
	return c.JSON(fiber.Map{"message": "Reward added"})
}

// AdminAddPartner creates a new partner account by the admin
func AdminAddPartner(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	if role!="admin"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	u := new(models.User)
	if err := c.BodyParser(u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	u.Password, _ = utils.HashingPassword(u.Password)
	u.Role = "partner"
	u.IsVerified = true
	db.DB.Create(u)
	return c.JSON(fiber.Map{"message": "Partner account created"})
}

// GetAllPartners retrieves all partners from the database
func GetAllPartners(c *fiber.Ctx) error {
	var partners []models.User
	if err := db.DB.Where("role = ?", "partner").Find(&partners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch partners"})
	}
	return c.JSON(partners)
}

// AdminUpdateReward updates an existing reward by the admin
func AdminUpdateReward(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idUint64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reward ID"})
	}
	rewardID := uint(idUint64)

	var reward models.Reward
	if err := db.DB.First(&reward, rewardID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Reward not found"})
	}

	if err := c.BodyParser(&reward); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := db.DB.Save(&reward).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update reward"})
	}

	return c.JSON(fiber.Map{"message": "Reward updated successfully", "reward": reward})
}

// AdminDeleteReward deletes an existing reward by the admin
func AdminDeleteReward(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idUint64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reward ID"})
	}
	rewardID := uint(idUint64)

	if err := db.DB.Delete(&models.Reward{}, rewardID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete reward"})
	}

	return c.JSON(fiber.Map{"message": "Reward deleted successfully"})
}

// GetAdminAnalytics provides a platform-wide overview for administrators.
func GetAdminAnalytics(c *fiber.Ctx) error {
	type ActiveMerchant struct {
	Username        string `json:"username"`
	RedemptionCount int    `json:"redemption_count"`
    }
	// Get total user and merchant counts
	var totalUsers int64
	db.DB.Model(&models.User{}).Where("role = ?", "user").Count(&totalUsers)

	var totalMerchants int64
	db.DB.Model(&models.User{}).Where("role = ?", "partner").Count(&totalMerchants)

	// Get total counts for rewards and redemptions
	var totalRewards int64
	db.DB.Model(&models.Reward{}).Count(&totalRewards)

	var totalRedemptions int64
	db.DB.Model(&models.Transaction{}).Count(&totalRedemptions)

	// Find the most active partners
	var mostActivePartners []ActiveMerchant
	db.DB.Model(&models.User{}).
		Select("users.username, COUNT(transactions.id) as redemption_count").
		Joins("JOIN rewards ON rewards.created_by_id = users.id").
		Joins("JOIN transactions ON transactions.reward_id = rewards.id").
		Where("users.role = ?", "partner").
		Group("users.username").
		Order("redemption_count DESC").
		Limit(5).
		Scan(&mostActivePartners)

	return c.JSON(fiber.Map{
		"total_users":           totalUsers,
		"total_merchants":       totalMerchants,
		"total_rewards":         totalRewards,
		"total_redemptions":     totalRedemptions,
		"most_active_partners": mostActivePartners,
	})
}

// PartnerAddReward adds a new reward created by the logged-in partner
func PartnerAddReward(c *fiber.Ctx) error {
	//Get Logged-in Partner
	role:=c.Locals("role").(string)
	userID:= uint(c.Locals("user_id").(float64))
	if role!="partner"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	r := new(models.Reward)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	r.CreatedByID = userID
	db.DB.Create(r)
	return c.JSON(fiber.Map{"message": "Partner reward added"})
}

// PartnerUpdateReward updates a reward created by the logged-in partner
func PartnerUpdateReward(c *fiber.Ctx) error {
	//Get Logged-in Partner
	role:=c.Locals("role").(string)
	userID:= uint(c.Locals("user_id").(float64))
	if role!="partner"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	//Get Reward ID from URL
	rewardID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid reward ID"})
	}

	//Find the Reward and Verify 
	var reward models.Reward
	if err := db.DB.First(&reward, rewardID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Reward not found"})
	}
	if reward.CreatedByID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You do not own this reward"})
	}
	var updatedData models.Reward
	if err := c.BodyParser(&updatedData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	// Update the fields
	if updatedData.Name != "" {
		reward.Name = updatedData.Name
	}
	if updatedData.Category != "" {
		reward.Category = updatedData.Category
	}
	if updatedData.Cost > 0 {
		reward.Cost = updatedData.Cost
	}
	if updatedData.Stock >= 0 {
		reward.Stock = updatedData.Stock
	}
	if updatedData.Discount > 0 {
		reward.Discount = updatedData.Discount
	}
	if updatedData.CampaignName != "" {
		reward.CampaignName = updatedData.CampaignName
	}
	if updatedData.Description != "" {
		reward.Description = updatedData.Description
	}
	db.DB.Save(&reward)

	return c.JSON(reward)
}

// DeleteReward deletes a reward created by the logged-in partner
func PartnerDeleteReward(c *fiber.Ctx) error {
	//Get Logged-in Partner
	role:=c.Locals("role").(string)
	userID:= uint(c.Locals("user_id").(float64))
	if role!="partner"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}

	//Get Reward ID from URL
	rewardID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid reward ID"})
	}

	// Find the Reward and Verify
	var reward models.Reward
	if err := db.DB.First(&reward, rewardID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Reward not found"})
	}
	if reward.CreatedByID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden: You do not own this reward"})
	}
	// Delete the Rewar
	db.DB.Delete(&reward)
	return c.Status(200).JSON(fiber.Map{"message": "Reward deleted successfully"})
}

// GetPartnerRewards retrieves all rewards created by the logged-in partner
func GetPartnerRewards(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	userID:= uint(c.Locals("user_id").(float64))
	if role!="partner"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
    var rewards []models.Reward
    db.DB.Where("created_by_id = ?", userID).Find(&rewards)
    return c.JSON(rewards)
}

// GetPartnerAnalytics retrieves analytics for the logged-in partner
func GetPartnerAnalytics(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	userID:= uint(c.Locals("user_id").(float64))
	if role!="partner"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	var rewards []models.Reward
	db.DB.Where("created_by_id = ?", userID).Find(&rewards)
	var totalRedemptions int64
	err:=db.DB.Model(&models.Transaction{}).
		Joins("JOIN rewards ON rewards.id = transactions.reward_id").
		Where("rewards.created_by_id = ?", userID).
		Count(&totalRedemptions)
	if err==nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch total redemptions"})
	}
	
	type PopularReward struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	//Get the most popular rewards for the partner
	var popularRewards []PopularReward
	db.DB.Model(&models.Transaction{}).
		Select("rewards.name, count(transactions.id) as count").
		Joins("JOIN rewards ON rewards.id = transactions.reward_id").
		Where("rewards.created_by_id = ?", userID).
		Group("rewards.name").
		Order("count desc").
		Limit(5).
		Scan(&popularRewards)

	return c.JSON(fiber.Map{
		"total_redemptions":    totalRedemptions,
		"most_popular_rewards": popularRewards,
	})
}

// ViewProfile retrieves the profile of the logged-in user
func ViewProfile(c *fiber.Ctx) error {
	userID:= uint(c.Locals("user_id").(float64))
	var u models.User
	db.DB.First(&u, userID)
	return c.JSON(fiber.Map{"username" : u.Username, "points": u.Points})
}
