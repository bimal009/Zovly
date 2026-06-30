import { defineRelations } from "drizzle-orm";
import * as schema from "./schema";

export const relations = defineRelations(schema, (r) => ({
	account: {
		user: r.one.user({
			from: r.account.userId,
			to: r.user.id
		}),
	},
	user: {
		accounts: r.many.account(),
		businesses: r.many.business({
			alias: "business_id_user_id_via_businessMembers"
		}),
		memberInvitesInvitedById: r.many.memberInvites({
			alias: "memberInvites_invitedById_user_id"
		}),
		memberInvitesRevokedById: r.many.memberInvites({
			alias: "memberInvites_revokedById_user_id"
		}),
		sessions: r.many.session(),
		business: r.one.business({
			from: r.user.businessId,
			to: r.business.id,
			alias: "user_businessId_business_id"
		}),
	},
	appConnections: {
		business: r.one.business({
			from: r.appConnections.businessId,
			to: r.business.id
		}),
	},
	business: {
		appConnections: r.many.appConnections(),
		appCredentials: r.many.appCredentials(),
		usersViaBusinessMembers: r.many.user({
			from: r.business.id.through(r.businessMembers.businessId),
			to: r.user.id.through(r.businessMembers.userId),
			alias: "business_id_user_id_via_businessMembers"
		}),
		plans: r.many.plans({
			from: r.business.id.through(r.businessSubscriptions.businessId),
			to: r.plans.id.through(r.businessSubscriptions.planId)
		}),
		categoriesBusinessId: r.many.categories({
			alias: "categories_businessId_business_id"
		}),
		productsViaConversations: r.many.products({
			alias: "products_id_business_id_via_conversations"
		}),
		faqs: r.many.faqs(),
		knowledgeChunks: r.many.knowledgeChunks(),
		memberInvites: r.many.memberInvites(),
		messageEmbeddings: r.many.messageEmbeddings(),
		conversations: r.many.conversations({
			from: r.business.id.through(r.messages.businessId),
			to: r.conversations.id.through(r.messages.conversationId)
		}),
		paymentRecords: r.many.paymentRecords(),
		policies: r.many.policies(),
		productsViaProductVariants: r.many.products({
			from: r.business.id.through(r.productVariants.businessId),
			to: r.products.id.through(r.productVariants.productId),
			alias: "business_id_products_id_via_productVariants"
		}),
		categoriesViaProducts: r.many.categories({
			from: r.business.id.through(r.products.businessId),
			to: r.categories.id.through(r.products.categoryId),
			alias: "business_id_categories_id_via_products"
		}),
		services: r.many.services(),
		usersBusinessId: r.many.user({
			alias: "user_businessId_business_id"
		}),
	},
	appCredentials: {
		business: r.one.business({
			from: r.appCredentials.businessId,
			to: r.business.id
		}),
	},
	plans: {
		businesses: r.many.business(),
		paymentRecords: r.many.paymentRecords(),
	},
	categories: {
		business: r.one.business({
			from: r.categories.businessId,
			to: r.business.id,
			alias: "categories_businessId_business_id"
		}),
		businesses: r.many.business({
			alias: "business_id_categories_id_via_products"
		}),
	},
	products: {
		businessesViaConversations: r.many.business({
			from: r.products.id.through(r.conversations.activeProductId),
			to: r.business.id.through(r.conversations.businessId),
			alias: "products_id_business_id_via_conversations"
		}),
		businessesViaProductVariants: r.many.business({
			alias: "business_id_products_id_via_productVariants"
		}),
	},
	faqs: {
		business: r.one.business({
			from: r.faqs.businessId,
			to: r.business.id
		}),
	},
	knowledgeChunks: {
		business: r.one.business({
			from: r.knowledgeChunks.businessId,
			to: r.business.id
		}),
	},
	memberInvites: {
		business: r.one.business({
			from: r.memberInvites.businessId,
			to: r.business.id
		}),
		userInvitedById: r.one.user({
			from: r.memberInvites.invitedById,
			to: r.user.id,
			alias: "memberInvites_invitedById_user_id"
		}),
		userRevokedById: r.one.user({
			from: r.memberInvites.revokedById,
			to: r.user.id,
			alias: "memberInvites_revokedById_user_id"
		}),
	},
	messageEmbeddings: {
		business: r.one.business({
			from: r.messageEmbeddings.businessId,
			to: r.business.id
		}),
		conversation: r.one.conversations({
			from: r.messageEmbeddings.conversationId,
			to: r.conversations.id
		}),
		message: r.one.messages({
			from: r.messageEmbeddings.messageId,
			to: r.messages.id
		}),
	},
	conversations: {
		messageEmbeddings: r.many.messageEmbeddings(),
		businesses: r.many.business(),
	},
	messages: {
		messageEmbeddings: r.many.messageEmbeddings(),
	},
	paymentRecords: {
		business: r.one.business({
			from: r.paymentRecords.businessId,
			to: r.business.id
		}),
		plan: r.one.plans({
			from: r.paymentRecords.planId,
			to: r.plans.id
		}),
		businessSubscription: r.one.businessSubscriptions({
			from: r.paymentRecords.subscriptionId,
			to: r.businessSubscriptions.id
		}),
	},
	businessSubscriptions: {
		paymentRecords: r.many.paymentRecords(),
	},
	policies: {
		business: r.one.business({
			from: r.policies.businessId,
			to: r.business.id
		}),
	},
	services: {
		business: r.one.business({
			from: r.services.businessId,
			to: r.business.id
		}),
	},
	session: {
		user: r.one.user({
			from: r.session.userId,
			to: r.user.id
		}),
	},
}))