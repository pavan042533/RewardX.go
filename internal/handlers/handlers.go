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

func LoginHandler(c *fiber.Ctx) error {
    var credentials struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.BodyParser(&credentials); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }
    if credentials.Username == "" && credentials.Email == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username or Email required"})
    }
    var user models.User
    result := db.DB.Where("username = ? OR email = ?", credentials.Username, credentials.Email).First(&user)
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

func ListRewards(c *fiber.Ctx) error {
	rewards:= []models.Reward{}
	if result:= db.DB.Find(&rewards); result.Error!=nil{
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Could not find Rewards"})
	}
	return c.JSON(rewards)
}

func GetUserWallet(c*fiber.Ctx) error {
	userID:= uint(c.Locals("user_id").(float64))
	var user models.User
	if result:= db.DB.Where("id = ?", userID).First(&user); result.Error!=nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(fiber.Map{"points": user.Points})
}

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

func GetUserTransactions(c *fiber.Ctx) error {
	userID:=uint(c.Locals("user_id").(float64))
	var transactions []models.Transaction
	db.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&transactions)
	return c.JSON(transactions)
}

func AdminAddReward(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	if role!="admin"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	var reward models.Reward
	if err:=c.BodyParser(&reward); err!=nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	db.DB.Create(&reward)
	return c.JSON(fiber.Map{"message": "Reward added"})
}

func AdminAddPartner(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	if role!="admin"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	u := new(models.User)
	if err := c.BodyParser(u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	u.Role = "partner"
	u.IsVerified = true
	db.DB.Create(u)
	return c.JSON(fiber.Map{"message": "Partner account created"})
}

func GetAllPartners(c *fiber.Ctx) error {
	var partners []models.User
	if err := db.DB.Where("role = ?", "partner").Find(&partners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch partners"})
	}
	return c.JSON(partners)
}

func UpdateReward(c *fiber.Ctx) error {
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

func DeleteReward(c *fiber.Ctx) error {
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

func PartnerAddReward(c *fiber.Ctx) error {
	role:=c.Locals("role").(string)
	if role!="partner"{
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid Access"})
	}
	r := new(models.Reward)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.DB.Create(r)
	return c.JSON(fiber.Map{"message": "Partner reward added"})
} 


func ViewProfile(c *fiber.Ctx) error {
	userID:= uint(c.Locals("user_id").(float64))
	var u models.User
	db.DB.First(&u, userID)
	return c.JSON(fiber.Map{"username" : u.Username, "points": u.Points})
}

