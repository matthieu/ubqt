{-# LANGUAGE TypeSynonymInstances #-}

module Main (main) where

import System.Environment ( getArgs )
import MonadUtils

import GHC
import Outputable
import GHC.Paths ( libdir )
import DynFlags ( defaultDynFlags )

import Type
import HscTypes ( cm_binds, CoreModule(..), mkTypeEnv, HscEnv(..) )
import CoreSyn
import IdInfo
import Var as V
import Name as N
import OccName as O
import DriverPipeline as DP
import DriverPhases as DPh

import PrelNames as PN
import TysPrim as TP
import TysWiredIn as TW

import Unique as U
import SrcLoc as SL
import FastString as FS

main = getArgs >>= doGHCStuff . head >>= putStrLn

instance Show CoreExpr where
  show (Var id)   = showVar id
  show (Lit l)    = "(Lit " ++ show l ++ ")"
  show (App e a)  = "(App " ++ show e ++ " " ++ show a ++ ")"
  show (Lam v e)  = "(Lam " ++ showVar v ++ " " ++ show e ++ ")"
  show (Let b e)  = "(Let " ++ show b ++ " " ++ show e ++ ")"
  show (Case e b t a) = "(Case " ++ show e ++ " " ++ show b ++ " T " ++ show a ++ ")"
  show (Cast e c) = "(Cast " ++ show e ++ " Coercion(" ++ show c  ++ ")" ++ ")"
  show (Note n e) = "(Note " ++ "N" ++ " " ++ show e ++ ")"
  show (Type t)   = "(Type " ++ show t ++ ")"
 
showVar v = "(Var details=" ++ (showIdDetails . V.idDetails) v ++ ",name=" ++ (showName . V.varName) v
              ++ ",type=" ++ show (V.varType v)
              ++ ",local=" ++ (show . isLocalVar) v ++ ",global=" ++ (show . isGlobalId) v ++ ",expo=" ++ (show . isExportedId) v ++ ")"

showName n = "{" ++ (moduleNameString . moduleName . nameModule) n ++ "." ++ (N.occNameString . N.nameOccName) n
              ++ ":isSys=" ++ (show . N.isSystemName) n 
              ++ ",isInt=" ++ (show . N.isInternalName) n
              ++ ",isExt=" ++ (show . N.isExternalName) n
              ++ ",isWir=" ++ (show . N.isWiredInName) n ++ "}"

showIdDetails VanillaId = "Vanilla"
showIdDetails (PrimOpId po) = "PrimOp " ++ show po
showIdDetails (ClassOpId co) = "ClassOp " ++ show co
showIdDetails (FCallId fc) = "Foreign"

instance Show CoreBind where
  show (NonRec b expr) = "NonRec " ++ (showVar b) ++ " " ++ (show expr)
  show (Rec arr) = "Rec [" ++ showRecArr arr ++ "]"

showRecArr []            = ""
showRecArr ((v,expr):bs) = "(" ++ showVar v ++ "," ++ show expr ++ ")," ++ showRecArr bs

instance Show Type where
  show = showSDoc . pprType

doGHCStuff f =
  defaultErrorHandler defaultDynFlags $ do
    runGhc (Just libdir) $ do
      dflags <- getSessionDynFlags
      setSessionDynFlags dflags

      coreMod <- compileToCoreModule f

      ioTyCon   <- lookupTyCon ioTyConName
      runMainIO <- lookupId PN.runMainIOName

      return ((show . cm_binds) coreMod ++ "\n\n" ++ show [mainNR ioTyCon, runMainIONR ioTyCon runMainIO])

lookupTyCon name = do
  tyCon <- lookupName name
  case tyCon of
    Just (ATyCon tc) -> return tc
    _                -> error "No name found"

lookupId name = do
  tyThing <- lookupName name
  case tyThing of
    Just (AnId id)  -> return id
    _               -> error "No name found"

coreMod ioTyCon runMainIO = CoreModule PN.mAIN (mkTypeEnv []) [mainNR ioTyCon, runMainIONR ioTyCon runMainIO] []

mainNR ioTyCon = NonRec (mainVar ioTyCon) (App (Var $ putStrLnVar ioTyCon) (App (Var unpackCStringVar) (mkStringLit "Hello")))

runMainIONR ioTyCon runMainIO = NonRec (rootMainVar ioTyCon) (App (App (Var runMainIO) (Type TW.unitTy)) (Var $ mainVar ioTyCon))

mainVar ioTyCon = globaliseId $ mkExportedLocalVar VanillaId (mkExtName PN.mAIN "main" 1) (mkTyConApp ioTyCon [TW.unitTy]) vanillaIdInfo

rootMainVar ioTyCon = globaliseId $ mkExportedLocalVar VanillaId 
                                      (N.mkExternalName PN.rootMainKey PN.rOOT_MAIN (mkVarOccFS (fsLit "main")) mkNoSrcSpan) 
                                      (mkTyConApp ioTyCon [TW.unitTy]) vanillaIdInfo

putStrLnVar ioTyCon = globaliseId $ 
  mkExportedLocalVar VanillaId (mkExtName PN.sYSTEM_IO "putStrLn" 2) (mkFunTy TW.stringTy $ mkTyConApp ioTyCon [TW.unitTy]) vanillaIdInfo

unpackCStringVar = globaliseId $ mkExportedLocalVar VanillaId PN.unpackCStringName (mkFunTy TP.addrPrimTy TW.stringTy) vanillaIdInfo

-- TODO use UniqSupply
mkExtName mod def uniq = N.mkExternalName (U.mkPseudoUniqueH uniq) mod (O.mkVarOccFS $ FS.fsLit def) mkNoSrcSpan  

mkNoSrcSpan = SL.mkSrcSpan (SL.mkSrcLoc (FS.fsLit "<prog>") 0 0) (SL.mkSrcLoc (FS.fsLit "<prog>") 0 0)

